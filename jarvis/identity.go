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
	"shanhu.io/g/settings"
)

type identityStore struct {
	settings settings.Settings
	key      string
}

func newIdentityStore(s settings.Settings, k string) *identityStore {
	return &identityStore{settings: s, key: k}
}

func (s *identityStore) Load(v interface{}) error {
	return s.settings.Get(s.key, v)
}

func (s *identityStore) Check() (bool, error) {
	return s.settings.Has(s.key)
}

func (s *identityStore) Save(v interface{}) error {
	return s.settings.Set(s.key, v)
}
