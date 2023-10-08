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

package apputil

import (
	"testing"

	"shanhu.io/g/pisces"
	"shanhu.io/g/settings"
)

func TestReadPassword(t *testing.T) {
	tables := pisces.NewTables(nil)
	s := settings.NewTable(tables)

	const key = "password"
	pwd, err := ReadPasswordOrSetRandom(s, key)
	if err != nil {
		t.Fatal("set password: ", err)
	}

	again, err := ReadPasswordOrSetRandom(s, key)
	if err != nil {
		t.Fatal("read password: ", err)
	}
	if pwd != again {
		t.Errorf("password got %q, want/was %q", again, pwd)
	}
}
