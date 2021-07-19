// Copyright (C) 2021  Shanhu Tech Inc.
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

package jarvis

import (
	"shanhu.io/misc/errcode"
)

func cmdUpdateGrubConfig(args []string) error {
	flags := cmdFlags.New()
	dev := flags.String("dev", "/dev/sda1", "boot partition device")
	osVersion := flags.String("os", "burmilla/os:v1.9.1", "os version")
	args = flags.ParseArgs(args)
	if len(args) != 0 {
		return errcode.InvalidArgf("expects no args")
	}
	return updateBootPart(*dev, *osVersion)
}
