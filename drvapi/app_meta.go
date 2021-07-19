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

package drvapi

// StepVersion records a particular version of a nextcloud or postgres release.
type StepVersion struct {
	Major   int
	Version string
	Source  string `json:",omitempty"`
	Image   string
}

// AppMeta stores the meta information of an HomeDrive application
type AppMeta struct {
	Name string

	// Dependencies.
	Deps []string `json:",omitempty"`

	// Version counter. To prevent rolling back. Most apps
	// do not support rolling back.
	Version int64 `json:",omitempty"`

	// SemVersion tracks compability.
	SemVersion string `json:",omitempty"`

	// Image ID, for simple single container apps.
	Image string `json:",omitempty"`

	// Steps is for apps that needs an upgrade ladder.
	Steps []*StepVersion `json:",omitempty"`
}
