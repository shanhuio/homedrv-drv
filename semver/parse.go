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
	"strconv"
	"strings"

	"shanhu.io/g/errcode"
)

// Major returns the major version number of a version string.
func Major(v string) (int, error) {
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return 0, errcode.InvalidArgf("invalid version: %q", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, errcode.Annotatef(err, "parse major version: %q", v)
	}
	if major <= 0 {
		return 0, errcode.InvalidArgf("invalid major version in %q", v)
	}
	return major, nil
}
