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

// Package homeboot provides a command line tool that creates a private key
// identity and enrolls it.
package homeboot

import (
	"shanhu.io/misc/subcmd"
)

// Main is the main entrance of the command line.
func Main() {
	c := subcmd.New()
	c.Add("install", "installs a new homedrive", cmdInstall)
	c.Add("uninstall", "uninstalls homedrive installation", cmdUninstall)
	c.Add(
		"cloud-config", "prints cloud-config for a new homedrive",
		cmdCloudConfig,
	)
	c.Add("enroll", "manually enroll an endpoint using passcode", cmdEnroll)
	c.Add("serve", "serves a local http service for installation", cmdServe)

	c.Main()
}
