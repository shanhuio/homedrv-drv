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
	"shanhu.io/homedrv/drvapi"
)

func makeManifest(metas []*drvapi.AppMeta) map[string]*drvapi.AppMeta {
	m := make(map[string]*drvapi.AppMeta)
	for _, meta := range metas {
		m[meta.Name] = meta
	}
	return m
}

func sameAppVersion(m1, m2 *drvapi.AppMeta) bool {
	return m1.Version == m2.Version &&
		m1.SemVersion == m2.SemVersion &&
		m1.Image == m2.Image
}

func appImage(meta *drvapi.AppMeta) string {
	if meta.Image != "" {
		return meta.Image
	}
	if n := len(meta.Steps); n > 0 {
		last := meta.Steps[n-1]
		return last.Image
	}
	return ""
}

type appQuerier interface {
	// Returns the meta info for an app. Returns NotFound error if app not
	// found.
	meta(name string) (*drvapi.AppMeta, error)
}
