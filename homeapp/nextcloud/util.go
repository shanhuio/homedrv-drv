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
	"io"
	"strings"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/homedrv/drv/executil"
)

func configSaysInstalled(config []byte) bool {
	// check if the config file has the `'installed' => true` line.
	lines := strings.Split(string(config), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == `'installed' => true,` {
			return true
		}
		if line == `'installed' => false,` {
			return false
		}
	}
	return false
}

func exec(c *dock.Cont, cmd []string, out io.Writer) error {
	return executil.RetError(c.ExecWithSetup(&dock.ExecSetup{
		Cmd:    cmd,
		Stdout: out,
	}))
}

func aptUpdate(c *dock.Cont, out io.Writer) error {
	cmd := []string{"apt-get", "update"}
	return exec(c, cmd, out)
}

func aptInstall(c *dock.Cont, pkgs []string, out io.Writer) error {
	cmd := []string{"apt-get", "install", "-y"}
	cmd = append(cmd, pkgs...)
	return exec(c, cmd, out)
}

func enableSMB(c *dock.Cont, out io.Writer) error {
	peclList := new(bytes.Buffer)
	if err := exec(c, []string{"pecl", "list"}, peclList); err != nil {
		return errcode.Annotate(err, "pecl list")
	}
	lines := strings.Split(peclList.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "smbclient" {
			// smbclient already installed; let's skip.
			return nil
		}
	}

	cmd := []string{"pecl", "install", "smbclient"}
	if err := exec(c, cmd, out); err != nil {
		return errcode.Annotate(err, "pecl install")
	}
	cmd = []string{"docker-php-ext-enable", "smbclient"}
	if err := exec(c, cmd, out); err != nil {
		return errcode.Annotate(err, "docker-php-ext-enable")
	}
	return nil
}

func occRet(
	c *dock.Cont, args []string, out io.Writer,
) (int, error) {
	cmd := append([]string{"php", "occ"}, args...)
	return c.ExecWithSetup(&dock.ExecSetup{
		Cmd:    cmd,
		User:   "www-data",
		Stdout: out,
	})
}

func occ(c *dock.Cont, args []string, out io.Writer) error {
	return executil.RetError(occRet(c, args, out))
}

func occOutput(c *dock.Cont, args []string) ([]byte, error) {
	out := new(bytes.Buffer)
	if err := occ(c, args, out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func testReadConfig(cont *dock.Cont) ([]byte, error) {
	const configFile = "/var/www/html/config/config.php"

	ret, err := cont.ExecWithSetup(&dock.ExecSetup{
		Cmd:  []string{"/usr/bin/test", "-e", configFile},
		User: "www-data",
	})
	if err != nil {
		return nil, errcode.Annotate(err, "test config.php")
	}
	if ret != 0 {
		return nil, nil
	}
	return dock.ReadContFile(cont, configFile)
}

func cron(cont *dock.Cont) error {
	return executil.RetError(cont.ExecWithSetup(&dock.ExecSetup{
		Cmd:    []string{"php", "cron.php"},
		User:   "www-data",
		Stdout: io.Discard,
	}))
}

func fixKey(major int) string {
	switch major {
	case 20:
		return Key20Fixed
	case 21:
		return Key21Fixed
	case 22:
		return Key22Fixed
	case 23:
		return Key23Fixed
	case 24:
		return Key24Fixed
	case 25:
		return Key25Fixed
	}
	return ""
}

func setRedisPassword(cont *dock.Cont, pwd string) error {
	// TODO(h8liu): should first check if redis password is incorrect.
	args := []string{
		"config:system:set", "-q",
		"--value=" + pwd,    // value
		"redis", "password", // key
	}
	return occ(cont, args, nil)
}

func setCronMode(cont *dock.Cont) error {
	args := []string{
		"config:app:set", "-q", "--value=cron",
		"core", "backgroundjobs_mode",
	}
	return occ(cont, args, nil)
}
