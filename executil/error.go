// Copyright (C) 2023  Shanhu Tech Inc.
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

package executil

import (
	"shanhu.io/g/errcode"
)

// RetError wraps the return value and the error. If err is not nil, it
// return err. When err is nil, if ret is not 0, it returns an internal
// error.
func RetError(ret int, err error) error {
	if err != nil {
		return err
	}
	if ret != 0 {
		return errcode.Internalf("exit value: %d", ret)
	}
	return nil
}
