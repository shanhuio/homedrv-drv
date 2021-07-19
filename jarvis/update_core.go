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
	"log"
	"time"

	"shanhu.io/homedrv/homeboot"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/dock"
)

func killOldCoreIfExist(d *drive) error {
	cont := dock.NewCont(d.dock, d.oldCore())
	ok, err := cont.Exists()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if err := cont.Drop(); err != nil {
		if rmError := cont.ForceRemove(); rmError != nil {
			log.Println("force remove old core: ", rmError)
		}
		return err
	}
	return nil
}

func restartAs(d *drive, img string) error {
	// This is normally not necessary; just to make sure rename will succeed.
	if err := killOldCoreIfExist(d); err != nil {
		return errcode.Annotate(err, "kill old core")
	}
	if err := dock.RenameCont(
		d.dock, d.core(), d.oldCore(),
	); err != nil {
		return errcode.Annotate(err, "rename core to old")
	}

	hasSysDock := true
	if err := homeboot.CheckSystemDock(); err != nil {
		if !errcode.IsNotFound(err) {
			return errcode.Annotate(err, "check system docker")
		}
		hasSysDock = false
	}

	config := &homeboot.CoreConfig{
		Drive:       d.config,
		Image:       img,
		BindSysDock: hasSysDock,
	}
	id, err := homeboot.StartCore(d.dock, config)
	if err != nil {
		return err
	}
	log.Printf("new core started as %q", id)

	for { // Waiting to be killed.
		time.Sleep(time.Hour)
	}
	// unreachable.
}

func updateCore(d *drive, img string) error {
	// This is used in self-update in background, so this must be using
	// volumes already.
	self, err := dock.InspectCont(d.dock, d.core())
	if err != nil {
		return errcode.Annotate(err, "inspect self")
	}
	if self.Image == img {
		return errSameImage // Already up-to-date, no need to do anything.
	}
	return restartAs(d, img)
}
