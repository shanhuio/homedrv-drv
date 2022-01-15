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

package jarvis

import (
	"shanhu.io/misc/errcode"
	"shanhu.io/pisces/settings"
)

func readPasswordOrSetRandom(
	s settings.Settings, k string,
) (string, error) {
	pwd, err := settings.String(s, k)
	if err != nil {
		if errcode.IsNotFound(err) {
			pwd := randPassword()
			if err := s.Set(k, pwd); err != nil {
				return "", errcode.Annotate(err, "set password")
			}
			return pwd, nil
		}
		return "", errcode.Annotate(err, "read password")
	}
	return pwd, nil
}
