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

// DomainMap is the domain mapping for one app.
type DomainMap struct {
	App string                  `json:",omitempty"`
	Map map[string]*DomainEntry `json:",omitempty"`
}

// DomainEntry is an entry for an application domain map.
type DomainEntry struct {
	Dest string
}

// Domains is a table that saves the application domain mapping.
type Domains interface {
	// Set sets the domain mapping of an application.
	Set(m *DomainMap) error

	// Clear clears the domain mapping of an application.
	Clear(app string) error
}
