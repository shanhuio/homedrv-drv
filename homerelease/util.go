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

package homerelease

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"shanhu.io/aries/creds"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/rand"
)

// MakeReleaseName makes a new release name.
func MakeReleaseName(typ string) (string, error) {
	ch := typ
	if typ == "dev" {
		u, err := creds.CurrentUser()
		if err != nil {
			return "", errcode.Annotate(err, "get current user")
		}
		ch = "dev-" + u
	}

	date := time.Now().Format("20060102")
	return fmt.Sprintf("%s-%s-%s", ch, date, rand.HexBytes(3)), nil
}

func filePath(base string, parts ...string) string {
	p := path.Join(parts...)
	return filepath.Join(base, filepath.FromSlash(p))
}
