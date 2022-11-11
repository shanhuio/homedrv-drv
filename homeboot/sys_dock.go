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
	"shanhu.io/pub/errcode"
	"shanhu.io/pub/osutil"
)

const systemDockSock = "/var/run/system-docker.sock"

// CheckSystemDock checks if system docker exists. It returns NotFound error if
// not.
func CheckSystemDock() error {
	ok, err := osutil.IsSock(systemDockSock)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.NotFoundf("system docker not found")
	}
	return nil
}
