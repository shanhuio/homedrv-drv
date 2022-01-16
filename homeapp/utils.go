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

package homeapp

import (
	"shanhu.io/homedrv/drvapi"
)

// Image returns the image of an app based on its meta information.
func Image(meta *drvapi.AppMeta) string {
	if meta.Image != "" {
		return meta.Image
	}
	if n := len(meta.Steps); n > 0 {
		last := meta.Steps[n-1]
		return last.Image
	}
	return ""
}
