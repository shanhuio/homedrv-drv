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

	"shanhu.io/homedrv/drv/homeapp"
	"shanhu.io/misc/errcode"
	"shanhu.io/pisces/settings"
	"shanhu.io/virgo/dock"
)

func loadDoorwayConfig(d *drive) (*doorwayConfig, error) {
	domain, err := settings.String(d.settings, homeapp.KeyMainDomain)
	if err != nil {
		return nil, errcode.Annotate(err, "read main domain")
	}
	fabricsServer, err := settings.String(d.settings, keyFabricsServerDomain)
	if err != nil {
		if errcode.IsNotFound(err) {
			fabricsServer = "" // Make sure the result is clear.
		} else {
			return nil, errcode.Annotate(err, "read fabrics server address")
		}
	}
	return &doorwayConfig{
		domain:        domain,
		fabricsServer: fabricsServer,
	}, nil
}

func updateDoorway(d *drive, img string) error {
	config, err := loadDoorwayConfig(d)
	if err != nil {
		return errcode.Annotate(err, "load config")
	}
	dw := newDoorway(d, config)
	return dw.update(img)
}

type taskRecreateDoorway struct {
	drive *drive
}

func (t *taskRecreateDoorway) run() error {
	log.Println("re-creating doorway.")

	d := t.drive
	c := dock.NewCont(d.dock, d.cont(nameDoorway))
	info, err := c.Inspect()
	if err != nil {
		return errcode.Annotate(err, "inspect current doorway")
	}
	// Force update current doorway to recreate the container.
	if err := updateDoorway(d, info.Image); err != nil {
		return errcode.Annotate(err, "recreate doorway")
	}
	return nil
}
