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

package jarvis

import (
	"time"

	"shanhu.io/misc/errcode"
)

const failureWindow = 3 * time.Minute

type recentFailures struct {
	Timestamps []int64 `json:",omitempty"`
}

func (r *recentFailures) update(now time.Time) {
	cut := now.Add(-failureWindow).UnixNano()
	ts := make([]int64, 0, 5)
	for _, t := range r.Timestamps {
		if t > cut {
			ts = append(ts, t)
		}
	}
	r.Timestamps = ts
}

func (r *recentFailures) count(now time.Time) int {
	r.update(now)
	return len(r.Timestamps)
}

func (r *recentFailures) add(now time.Time) int {
	r.update(now)
	r.Timestamps = append(r.Timestamps, now.UnixNano())
	return len(r.Timestamps)
}

func (r *recentFailures) clear() {
	r.Timestamps = nil
}

var errTooManyFailures = errcode.Unauthorizedf("too many recent failures")
