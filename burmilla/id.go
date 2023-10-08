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

package burmilla

import (
	"strconv"
	"strings"

	"shanhu.io/g/errcode"
)

func parseIDOutput(bs []byte) (int, error) {
	return strconv.Atoi(strings.TrimSpace(string(bs)))
}

// UserID returns the uid of a particular user.
func UserID(b *Burmilla, name string) (int, error) {
	out, err := b.ExecOutput([]string{"id", "-u", name})
	if err != nil {
		return 0, errcode.Annotate(err, "get user id")
	}
	id, err := parseIDOutput(out)
	if err != nil {
		return 0, errcode.Annotate(err, "parse user id")
	}
	return id, nil
}

// GroupID returns the gid of a particular user
func GroupID(b *Burmilla, name string) (int, error) {
	out, err := b.ExecOutput([]string{"id", "-g", name})
	if err != nil {
		return 0, errcode.Annotate(err, "get group id")
	}
	id, err := parseIDOutput(out)
	if err != nil {
		return 0, errcode.Annotate(err, "parse group id")
	}
	return id, nil
}
