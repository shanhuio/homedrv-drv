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

package postgres

import (
	"fmt"
	"log"
	"time"

	"shanhu.io/pub/errcode"
	"shanhu.io/pub/sqlx"
)

func createDB(db *sqlx.DB, name, pwd string) error {
	x := fmt.Sprintf("create database %s", name)
	if _, err := db.X(x); err != nil {
		return errcode.Annotate(err, "create db")
	}
	x = fmt.Sprintf("create role %s login", name)
	if pwd != "" {
		x += fmt.Sprintf(" password '%s'", pwd)
	}
	if _, err := db.X(x); err != nil {
		return errcode.Annotate(err, "set password")
	}
	return nil
}

func dropDB(db *sqlx.DB, name string) error {
	x := fmt.Sprintf("drop database if exists %s", name)
	if _, err := db.X(x); err != nil {
		return errcode.Annotate(err, "drop db")
	}
	x = fmt.Sprintf("drop role if exists %s", name)
	if _, err := db.X(x); err != nil {
		return errcode.Annotate(err, "drop role")
	}
	return nil
}

func waitDB(db *sqlx.DB, timeout time.Duration) error {
	ping := func() bool {
		err := db.Ping()
		if err != nil {
			log.Printf("wait db: %v", err)
		}
		return err == nil
	}

	if ping() {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			if ping() {
				return nil
			}
		case <-timer.C:
			return errcode.TimeOutf("wait db time out")
		}
	}
}
