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
	"time"

	"shanhu.io/pub/errcode"
)

// Uptime returns the uptime of the system.
func Uptime(b *Burmilla) (time.Duration, error) {
	line, err := b.ExecOutput([]string{"cat", "/proc/uptime"})
	if err != nil {
		return 0, errcode.Annotate(err, "query system uptime")
	}

	fields := strings.Fields(string(line))
	if len(fields) != 2 {
		return 0, errcode.Internalf(
			"system uptime line has %d fields, want 2", len(fields),
		)
	}

	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, errcode.Annotate(err, "parse uptime")
	}

	uptime := time.Duration(int64(secs*1e9)) * time.Nanosecond
	return uptime, nil
}
