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
	"testing"

	"shanhu.io/homedrv/drvapi"
)

func TestManifestFromRelease(t *testing.T) {
	var empty drvapi.Release
	m := manifestFromRelease(&empty)
	if len(m) != 0 {
		t.Errorf("got %v, want empty map", m)
	}

	m2 := manifestFromRelease(nil)
	if len(m2) != 0 {
		t.Errorf("got %v, want empty map", m)
	}
}
