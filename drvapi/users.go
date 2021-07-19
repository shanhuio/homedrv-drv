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

// UserSSHKey saves a user's SSH public key.
type UserSSHKey struct {
	ID        string
	PublicKey string

	TimeCreatedSec int64 // Timestamp in seconds.
	TimeCreated    int64 // Timestamp in seconds, legacy.
}

// UserSSHKeys wraps a list of user SSH keys.
type UserSSHKeys struct {
	Keys []*UserSSHKey `json:",omitempty"`
}

// UserSSHKeyLines is a list of user SSH keys.
type UserSSHKeyLines struct {
	Keys []string
}
