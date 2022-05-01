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

package homeboot

import (
	"net/url"

	"shanhu.io/homedrv/drv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/virgo/bosinit"
)

func fetchUserKeys(user string) ([]string, error) {
	c := &httputil.Client{
		Server: &url.URL{
			Scheme: "https",
			Host:   "www.homedrive.io",
		},
	}
	resp := new(drvapi.UserSSHKeyLines)
	if err := c.Call("/pubapi/user/sshkeys", user, resp); err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

// FetchSSHKeys fetches the SSH keys specified by the config.
func FetchSSHKeys(c *InitConfig) ([]string, error) {
	var lines []string
	if c.GitHubKeys != "" {
		keys, err := bosinit.FetchGitHubKeys(c.GitHubKeys)
		if err != nil {
			return nil, errcode.Annotate(err, "fetch github keys")
		}
		lines = append(lines, keys...)
	}

	if c.UserKeys != "" {
		keys, err := fetchUserKeys(c.UserKeys)
		if err != nil {
			return nil, errcode.Annotatef(
				err, "fetch ssh keys of %q", c.UserKeys,
			)
		}
		lines = append(lines, keys...)
	}

	return lines, nil
}
