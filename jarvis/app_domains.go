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

package jarvis

import (
	"sort"

	"shanhu.io/homedrv/drv/homeapp"
	"shanhu.io/pub/errcode"
	"shanhu.io/pub/pisces"
)

type appDomains struct {
	t *pisces.KV
}

func newAppDomains(b *pisces.Tables) *appDomains {
	return &appDomains{t: b.NewKV("app_domains")}
}

func (b *appDomains) Set(m *homeapp.DomainMap) error {
	if len(m.Map) == 0 {
		return b.Clear(m.App)
	}
	return b.t.Replace(m.App, m)
}

func (b *appDomains) Clear(app string) error {
	if err := b.t.Remove(app); err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (b *appDomains) list() ([]*homeapp.DomainMap, error) {
	var maps []*homeapp.DomainMap
	it := &pisces.Iter{
		Make: func() interface{} { return new(homeapp.DomainMap) },
		Do: func(_ string, v interface{}) error {
			maps = append(maps, v.(*homeapp.DomainMap))
			return nil
		},
	}
	if err := b.t.Walk(it); err != nil {
		return nil, err
	}
	sort.Slice(maps, func(i, j int) bool {
		return maps[i].App < maps[j].App
	})
	return maps, nil
}
