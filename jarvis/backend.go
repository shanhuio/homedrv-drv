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
	"log"
	"time"

	"shanhu.io/aries/identity"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/osutil"
	"shanhu.io/misc/rand"
	"shanhu.io/pisces"
	"shanhu.io/pisces/settings"
)

type backend struct {
	tables *pisces.Tables

	settings     *settings.Table
	identity     identity.Core
	users        *users
	securityLogs *securityLogs
	appDomains   *appDomains
}

func newBackend(file string) (*backend, error) {
	if file == "" {
		file = "var/jarvis.db"
	}

	dbExist, err := osutil.Exist(file)
	if err != nil {
		return nil, errcode.Annotate(err, "check database file exist")
	}

	tables, err := pisces.OpenSqlite3Tables(file)
	if err != nil {
		return nil, err
	}

	users := newUsers(tables)
	settings := settings.NewTable(tables)

	id := identity.NewSimpleCore(
		newIdentityStore(settings, keyIdentity), nil,
	)

	b := &backend{
		tables: tables,

		settings:     settings,
		identity:     id,
		users:        users,
		securityLogs: newSecurityLogs(tables),
		appDomains:   newAppDomains(tables),
	}

	if !dbExist {
		log.Print("initializing backend")
	}
	if err := backendInit(b); err != nil {
		return nil, errcode.Annotate(err, "init backend")
	}

	return b, nil
}

func backendInit(b *backend) error {
	if err := b.tables.CreateMissing(); err != nil {
		return errcode.Annotate(err, "create backend tables")
	}

	if ok, err := b.settings.Has(keySessionHMAC); err != nil {
		return errcode.Annotate(err, "check session hmac")
	} else if !ok {
		log.Println("create hmac session key")
		const hmacKeyLen = 16
		sessionHmac := rand.Letters(hmacKeyLen)
		if err := b.settings.Set(keySessionHMAC, sessionHmac); err != nil {
			return errcode.Annotate(err, "set session hmac key")
		}
	}

	if ok, err := b.users.has(rootUser); err != nil {
		return errcode.Annotate(err, "check root user")
	} else if !ok {
		const passwordLen = 16
		pwd := rand.Letters(passwordLen)
		if err := b.users.create(rootUser, pwd); err != nil {
			return errcode.Annotate(err, "create root user")
		}
		if err := b.settings.Set(keyJarvisPass, pwd); err != nil {
			return errcode.Annotate(err, "save jarvis init password")
		}
	}

	// TODO(h8liu): shorten this, and implement rotation.
	const tenYears = 10 * 356 * 24 * time.Hour
	if _, err := b.identity.Init(&identity.CoreConfig{
		Keys: []*identity.KeyConfig{{
			NotValidAfter: time.Now().Add(tenYears).Unix(),
		}},
	}); err != nil && err != identity.ErrAlreadyInitialized {
		return errcode.Annotate(err, "init identity")
	}

	return nil
}

func (b *backend) kernel() *kernel {
	return &kernel{
		settings:   b.settings,
		appDomains: b.appDomains,
	}
}
