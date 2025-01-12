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
	"sort"

	"shanhu.io/g/errcode"
	"shanhu.io/g/settings"
	"shanhu.io/homedrv/drv/drvapi"
)

type appsState struct {
	Metas    map[string]*drvapi.AppMeta `json:",omitempty"`
	Anchored map[string]bool            `json:",omitempty"`
}

func (s *appsState) setMeta(app string, m *drvapi.AppMeta) {
	if s.Metas == nil {
		s.Metas = make(map[string]*drvapi.AppMeta)
	}
	if m == nil {
		delete(s.Metas, app)
	} else {
		s.Metas[app] = m
	}
}

func (s *appsState) meta(app string) *drvapi.AppMeta {
	if s.Metas == nil {
		return nil
	}
	return s.Metas[app]
}

func (s *appsState) setAnchor(app string, b bool) {
	if s.Anchored == nil {
		s.Anchored = make(map[string]bool)
	}
	if !b {
		delete(s.Anchored, app)
	} else {
		s.Anchored[app] = true
	}
}

func (s *appsState) list() []string {
	var list []string
	for name := range s.Metas {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

func (s *appsState) semVersions() map[string]string {
	m := make(map[string]string)
	for name, meta := range s.Metas {
		m[name] = meta.SemVersion
	}
	return m
}

type appsStateStore interface {
	save(s *appsState) error
	load() (*appsState, error)
}

type appsStateSettings struct {
	settings settings.Settings
	key      string
}

func (s *appsStateSettings) save(state *appsState) error {
	return s.settings.Set(s.key, state)
}

func (s *appsStateSettings) load() (*appsState, error) {
	state := new(appsState)
	if err := s.settings.Get(s.key, state); err != nil {
		if errcode.IsNotFound(err) {
			return new(appsState), nil
		}
		return nil, err
	}
	return state, nil
}

func (s *appsStateSettings) has() (bool, error) {
	return s.settings.Has(s.key)
}
