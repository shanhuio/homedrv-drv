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

package doorway

import (
	"strings"
	"sync"
)

// HomeHost is the destination mapping that maps to doorway's internal
// administration server.
const HomeHost = "~"

const (
	hostNone = iota
	hostHome
	hostProxy
	hostRedirect
)

type hostEntry struct {
	host string
	typ  int
}

type hostMap interface {
	mapHost(from string) *hostEntry
}

type memHostMap struct {
	mu sync.RWMutex

	m map[string]*hostEntry
}

func newMemHostMap(m map[string]string) *memHostMap {
	entries := make(map[string]*hostEntry)

	for from, to := range m {
		if to == HomeHost {
			entries[from] = &hostEntry{
				typ: hostHome,
			}
		} else if strings.HasPrefix(to, "!") {
			entries[from] = &hostEntry{
				typ:  hostRedirect,
				host: strings.TrimPrefix(to, "!"),
			}
		} else {
			entries[from] = &hostEntry{
				typ:  hostProxy,
				host: to,
			}
		}
	}

	return &memHostMap{m: entries}
}

func (m *memHostMap) mapHost(from string) *hostEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	to, ok := m.m[from]
	if !ok {
		return nil
	}
	cp := *to
	return &cp
}

func hostMapToProxy(m hostMap, from string) string {
	entry := m.mapHost(from)
	if entry == nil {
		return ""
	}
	if entry.typ != hostProxy {
		return ""
	}
	return entry.host
}

func hostMapHas(m hostMap, from string) bool {
	return m.mapHost(from) != nil
}
