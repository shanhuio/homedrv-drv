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

package jarvis

import (
	"testing"

	"shanhu.io/g/dock"
)

func TestCheckDockerVersion(t *testing.T) {
	for _, v := range []string{
		"20.10.10",
		"20.10.10-debug",
		"20.10.10-rc",
		"20.10.22",
		"20.11.22",
		"21.0.0",
		"100.0.0",
	} {
		info := &dock.VersionInfo{Version: v}
		if err := checkDockerVersion(info); err != nil {
			t.Errorf("%q should not error, got %v", v, err)
		}
	}

	for _, v := range []string{
		"",
		"1",
		"a",
		"1.0",
		"19.10.9",
		"20.10.9-rc",
		"20.10.9-debug",
		"20.10.9",
	} {
		info := &dock.VersionInfo{Version: v}
		if err := checkDockerVersion(info); err == nil {
			t.Errorf("%q should error, got nil", v)
		}
	}

}
