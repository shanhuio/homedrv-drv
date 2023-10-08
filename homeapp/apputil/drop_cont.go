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

package apputil

import (
	"errors"
	"log"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
)

// ErrSameImage is returned when there is no image change.
var ErrSameImage = errors.New("same image")

// DropIfDifferent drops the container with name if the image is
// different. It returns ErrSameImage if the image is the same.
func DropIfDifferent(d *dock.Client, name, img string) error {
	c := dock.NewCont(d, name)
	info, err := c.Inspect()
	if err != nil {
		if errcode.IsNotFound(err) {
			log.Printf("container %q not found", name)
			return nil
		}
		return errcode.Annotatef(err, "inspect %s", name)
	}
	if info.Image == img {
		return ErrSameImage // nothing to update
	}
	if err := c.Drop(); err != nil {
		return errcode.Annotatef(err, "drop %s", name)
	}
	return nil
}

// DropIfExists drops the container if the container exists.  Otherwise, it
// prints a log line and do nothing.
func DropIfExists(cont *dock.Cont) error {
	exists, err := cont.Exists()
	if err != nil {
		return errcode.Annotatef(err, "check container exists")
	}
	if !exists {
		log.Printf("container %q does not exist; skip dropping", cont.ID())
		return nil
	}
	return cont.Drop()
}
