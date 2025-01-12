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

package semver

import (
	"testing"
)

func TestMajor(t *testing.T) {
	for _, test := range []struct {
		input   string
		want    int
		wantErr bool
	}{
		{input: "1.0.0", want: 1},
		{input: "30.4.3", want: 30},
		{input: "24.12.343.7", want: 24},
		{input: "1", want: 1},
		{input: "3.7", want: 3},
		{input: "", wantErr: true},
		{input: "a.b", wantErr: true},
		{input: "..........", wantErr: true},
	} {
		got, err := Major(test.input)
		if err != nil {
			if !test.wantErr {
				t.Errorf("Major(%q), got error: %s", test.input, err)
			}
			continue
		}

		if test.wantErr {
			t.Errorf("Major(%q), got %d, want error", test.input, got)
		} else if got != test.want {
			t.Errorf("Major(%q), got %d, want %d", test.input, got, test.want)
		}
	}
}
