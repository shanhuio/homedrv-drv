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
	"fmt"
	"strconv"
	"strings"

	"shanhu.io/g/dock"
)

func checkDockerVersion(v *dock.VersionInfo) error {
	s := v.Version
	fields := strings.Split(s, ".")
	if len(fields) < 3 {
		return fmt.Errorf("invalid version string: %q", s)
	}

	major, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("invalid major in version %q", s)
	}
	minor, err := strconv.Atoi(fields[1])
	if err != nil {
		return fmt.Errorf("invalid minor in version %q", s)
	}

	num := fields[2]
	if pre, _, ok := strings.Cut(fields[2], "-"); ok {
		num = pre
	}
	maintanence, err := strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("invalid maintanence in version %q", s)
	}

	if major > 20 {
		return nil
	}
	if major == 20 && minor > 10 {
		return nil
	}
	if major == 20 && minor == 10 && maintanence >= 10 {
		return nil
	}

	return fmt.Errorf(
		"docker version %q too low, requires at least 20.10.10", s,
	)
}
