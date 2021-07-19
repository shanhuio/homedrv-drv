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

package jarvis

import (
	"shanhu.io/misc/flagutil"
)

var cmdFlags = flagutil.NewFactory("jarvis")

type clientFlags struct {
	config string
	db     string
}

func newClientFlags(flags *flagutil.FlagSet) *clientFlags {
	c := new(clientFlags)
	flags.StringVar(&c.config, "config", "var/config.jsonx", "config file")
	flags.StringVar(&c.db, "db", "var/jarvis.db", "database file")
	return c
}

func newClientDrive(flags *clientFlags) (*drive, error) {
	c, err := readConfig(flags.config)
	if err != nil {
		return nil, err
	}
	b, err := newBackend(flags.db)
	if err != nil {
		return nil, err
	}

	d, err := newDrive(c, b.kernel())
	if err != nil {
		return nil, err
	}
	return d, nil
}
