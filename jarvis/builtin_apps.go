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
	"shanhu.io/homedrv/drv/homeapp"
	"shanhu.io/homedrv/drv/homeapp/nextcloud"
	"shanhu.io/homedrv/drv/homeapp/postgres"
	"shanhu.io/homedrv/drv/homeapp/redis"
	"shanhu.io/pub/errcode"
)

type builtInApps struct {
	stubs map[string]*appStub
}

func newBuiltInApps(c homeapp.Core) *builtInApps {
	m := make(map[string]*appStub)
	for _, a := range []struct {
		name string
		app  homeapp.App
	}{
		{name: "redis", app: redis.New(c)},
		{name: "postgres", app: postgres.New(c)},
		{name: "ncfront", app: nextcloud.NewFront(c)},
		{name: "nextcloud", app: nextcloud.New(c)},
	} {
		m[a.name] = &appStub{App: a.app}
	}

	return &builtInApps{stubs: m}
}

func (b *builtInApps) makeStub(name string) (*appStub, error) {
	a, ok := b.stubs[name]
	if ok {
		return a, nil
	}
	return nil, errcode.NotFoundf("app %q not found", name)
}
