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
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/flagutil"
	"shanhu.io/misc/osutil"
)

var cmdFlags = flagutil.NewFactory("jarvis")

type clientFlags struct {
	home string
}

func newClientFlags(flags *flagutil.FlagSet) *clientFlags {
	c := new(clientFlags)
	flags.StringVar(&c.home, "home", ".", "home directory")
	return c
}

func newClientDrive(flags *clientFlags) (*drive, error) {
	h, err := osutil.NewHome(flags.home)
	if err != nil {
		return nil, errcode.Annotate(err, "new home")
	}
	c, err := readConfig(h)
	if err != nil {
		return nil, errcode.Annotate(err, "read config")
	}
	b, err := newBackend(h)
	if err != nil {
		return nil, err
	}
	d, err := newDrive(c, b.kernel())
	if err != nil {
		return nil, err
	}
	return d, nil
}
