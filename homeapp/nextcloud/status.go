// Copyright (C) 2022  Shanhu Tech Inc.
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your
// option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package nextcloud

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
)

type status struct {
	Installed     bool   `json:"installed"`
	Version       string `json:"version"`
	VersionString string `json:"versionstring"`
}

func parseNextcloudStatus(out string) (*status, error) {
	lines := strings.Split(out, "\n")
	var theLine string
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
			theLine = s
			break
		}
	}

	if theLine == "" {
		return nil, errcode.InvalidArgf("status not found: %q", out)
	}

	status := new(status)
	if err := json.Unmarshal([]byte(theLine), status); err != nil {
		return nil, errcode.Annotate(err, "parse status")
	}
	return status, nil
}

func readStatus(c *dock.Cont) (*status, int, error) {
	out := new(bytes.Buffer)
	ret, err := occRet(c, []string{"status", "--output=json"}, out)
	if err != nil {
		return nil, 0, errcode.Annotate(err, "occ status")
	}
	if ret != 0 {
		return nil, ret, nil
	}
	status, err := parseNextcloudStatus(out.String())
	return status, 0, err
}

var errNotInstalled = errcode.Internalf("status not ready")

func checkInstalled(c *dock.Cont, v string) error {
	// Double check with occ status. Should be safe now.
	status, ret, err := readStatus(c)
	if err != nil {
		return err
	}
	if ret != 0 {
		log.Printf("status exit with: %d", ret)
		return errNotInstalled
	}
	if v != "" && status.VersionString != v {
		log.Printf(
			"not correct version: want %q, got %q",
			v, status.VersionString,
		)
		return errNotInstalled
	}
	if !status.Installed {
		return errNotInstalled
	}
	return nil
}

func waitReady(
	cont *dock.Cont, timeout time.Duration, v string,
) error {
	start := time.Now()
	deadline := start.Add(timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Always give nextcloud 2 minute for it to do its upgrade / bootup
	// thing.
	log.Println("wait for 2 minute for nextcloud container to start")
	time.Sleep(2 * time.Minute)

	i := 1
	for range ticker.C {
		now := time.Now()
		if now.After(deadline) {
			return errcode.TimeOutf(
				"nextcloud install timeout in %s", timeout,
			)
		}

		if i%5 == 1 {
			dur := now.Sub(start)
			log.Printf("check nextcloud status (%.1f sec)", dur.Seconds())
		}
		i++

		config, err := testReadConfig(cont)
		if err != nil {
			return errcode.Annotate(err, "check config")
		}
		if config == nil {
			log.Printf("config.php not created yet")
			continue
		}
		if !configSaysInstalled(config) {
			continue
		}

		if err := checkInstalled(cont, v); err != nil {
			if err == errNotInstalled {
				continue
			}
			return errcode.Annotate(err, "check installed")
		}

		mode, err := occOutput(
			cont, []string{"maintenance:mode", "-n"},
		)
		if err != nil {
			return errcode.Annotatef(err, "check maintenance mode")
		}
		modeLine := strings.TrimSpace(string(mode))
		if !strings.HasSuffix(modeLine, " disabled") {
			log.Println("nextcloud in maintenance mode")
			continue
		}

		break // passed all checks
	}

	durSecs := time.Since(start).Seconds()
	if v == "" {
		log.Printf("nextcloud installed in %.1f second(s)", durSecs)
	} else {
		log.Printf("nextcloud %s installed in %.1f seconds(s)", v, durSecs)
	}
	return nil
}

func readTrueVersion(cont *dock.Cont) (string, error) {
	status, ret, err := readStatus(cont)
	if err != nil {
		return "", errcode.Annotate(err, "read status")
	}
	if ret != 0 {
		return "", errcode.Internalf("status exit with: %d", ret)
	}
	return status.VersionString, nil
}

func readVersion(cont *dock.Cont, def string) (string, error) {
	exists, err := cont.Exists()
	if err != nil {
		return "", errcode.Annotatef(err, "check container exist")
	}
	if !exists {
		return def, nil
	}

	// If container exists, try to get from true source.
	v, err := readTrueVersion(cont)
	if err != nil {
		log.Print("failed to read nextcloud version: ", err)
		return def, nil
	}
	return v, nil
}
