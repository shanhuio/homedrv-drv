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

	"shanhu.io/homedrv/executil"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/dock"
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

func aptUpdate(c *dock.Cont, out io.Writer) error {
	cmd := []string{"apt-get", "update"}
	return executil.RetError(c.ExecWithSetup(&dock.ExecSetup{
		Cmd:    cmd,
		Stdout: out,
	}))
}

func aptInstall(c *dock.Cont, pkg string, out io.Writer) error {
	cmd := []string{"apt-get", "install", "-y", pkg}
	return executil.RetError(c.ExecWithSetup(&dock.ExecSetup{
		Cmd:    cmd,
		Stdout: out,
	}))
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
	}
	return ""
}