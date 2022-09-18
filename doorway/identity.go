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

package doorway

import (
	"context"
	"os"
)

// Identity provides an identity for dialing fabrics.
type Identity interface {
	// Load loads the identity private key. Returns errcode.NotFound error
	// if key is not yet provisioned.
	Load(ctx context.Context) ([]byte, error)
}

type staticIdentity struct {
	pri []byte
}

func newStaticIdentity(bs []byte) *staticIdentity {
	return &staticIdentity{bs}
}

func newFileIdentity(f string) (*staticIdentity, error) {
	bs, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return newStaticIdentity(bs), nil
}

func (s *staticIdentity) Load(ctx context.Context) ([]byte, error) {
	cp := make([]byte, len(s.pri))
	copy(cp, s.pri)
	return cp, nil
}

// NewFileIdentity loads a private key from a file.
func NewFileIdentity(f string) (Identity, error) {
	id, err := newFileIdentity(f)
	if err != nil {
		return nil, err
	}
	return id, nil
}
