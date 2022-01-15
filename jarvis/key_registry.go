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
	"shanhu.io/aries"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/rsautil"
)

type keyRegistry struct {
	users *users
}

func newKeyRegistry(users *users) *keyRegistry {
	return &keyRegistry{users: users}
}

func (r *keyRegistry) set(user string, keyBytes []byte) error {
	if user != rootUser {
		return errcode.InvalidArgf("only root user supported")
	}
	return r.users.mutate(user, func(info *userInfo) error {
		info.APIKeys = keyBytes
		return nil
	})
}

func (r *keyRegistry) Keys(user string) ([]*rsautil.PublicKey, error) {
	if user != rootUser {
		return nil, errcode.InvalidArgf("only root user supported")
	}

	info, err := r.users.get(user)
	if err != nil {
		return nil, errcode.Annotate(err, "get user info")
	}
	return rsautil.ParsePublicKeys(info.APIKeys)
}

func (r *keyRegistry) apiSet(c *aries.C, keyBytes []byte) error {
	return r.set(rootUser, keyBytes)
}
