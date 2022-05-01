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
	"shanhu.io/homedrv/drv/drvapi"
	"shanhu.io/homedrv/drv/homeapp/nextcloud"
	"shanhu.io/homedrv/drv/homeapp/postgres"
	"shanhu.io/homedrv/drv/homeapp/redis"
	"shanhu.io/misc/errcode"
)

func appsStateForLegacyUpgrade(reg *appRegistry) (
	*appsState, error,
) {
	state := &appsState{
		Metas:    make(map[string]*drvapi.AppMeta),
		Anchored: map[string]bool{nextcloud.Name: true},
	}
	for _, name := range []string{
		redis.Name,
		postgres.Name,
		nextcloud.Name,
		nextcloud.NameFront,
	} {
		meta, err := reg.meta(name)
		if err != nil {
			return nil, err
		}
		state.Metas[name] = meta
	}
	return state, nil
}

func maybeSetAppsStateFromLegacy(
	s *appsStateSettings, reg *appRegistry,
) error {
	hasState, err := s.has()
	if err != nil {
		return errcode.Annotate(err, "check apps state")
	}
	if hasState {
		return nil
	}

	// no state yet, and we have nextcloud, so this is an upgrade
	// from legacy system.
	state, err := appsStateForLegacyUpgrade(reg)
	if err != nil {
		return errcode.Annotate(err, "build apps state")
	}
	return s.save(state)
}
