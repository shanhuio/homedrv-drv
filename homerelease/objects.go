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

package homerelease

import (
	"os"
	"sort"

	"shanhu.io/misc/errcode"
	"shanhu.io/misc/tarutil"
)

func writeObjects(p string, objects map[string]string) error {
	var sums []string
	for sum := range objects {
		sums = append(sums, sum)
	}
	sort.Strings(sums)

	stream := tarutil.NewStream()
	for _, sum := range sums {
		stream.AddFile(sum, tarutil.ModeMeta(0644), objects[sum])
	}

	f, err := os.Create(p)
	if err != nil {
		return errcode.Annotate(err, "create file")
	}
	defer f.Close()

	if _, err := stream.WriteTo(f); err != nil {
		return errcode.Annotate(err, "write tarball")
	}
	if err := f.Sync(); err != nil {
		return errcode.Annotate(err, "sync to disk")
	}
	return nil
}
