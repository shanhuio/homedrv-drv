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

package drvapi

import (
	"testing"
)

func TestChannelName(t *testing.T) {
	for _, test := range []struct {
		name string
		base string
		arch string
	}{
		{"stable", "stable", "amd64"},
		{"alpha", "alpha", "amd64"},
		{"stable-amd64", "stable", "amd64"},
		{"stable-arm64", "stable", "arm64"},
		{"alpha-arm64", "alpha", "arm64"},
	} {
		parsed := ParseChannelName(test.name)
		if parsed.Base != test.base {
			t.Errorf(
				"parse channel name: got base %q, want %q",
				parsed.Base, test.base,
			)
		}
		if got := parsed.Architecture(); got != test.arch {
			t.Errorf(
				"parse channel name: got arch %q, want %q",
				got, test.arch,
			)
		}
		if got := parsed.String(); got != test.name {
			t.Errorf(
				"parsed channel name: got %q, want %q",
				got, test.name,
			)
		}
	}
}
