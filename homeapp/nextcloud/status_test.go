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

package nextcloud

import (
	"testing"
)

func TestParseNextcloudStatus(t *testing.T) {
	const s = "Nextcloud is weird...\n" +
		`{"installed": false, "version": "0.1"}`

	status, err := parseNextcloudStatus(s)
	if err != nil {
		t.Errorf("parse %q: %s", s, err)
	}
	if status.Installed != false {
		t.Error("want not installed")
	}
	if want := "0.1"; status.Version != want {
		t.Errorf("wrong version: got %q, want %q", status.Version, want)
	}
}
