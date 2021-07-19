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
	"shanhu.io/homedrv/drvapi"
)

type app interface {
	// Called when the version is changed from a non-empty string to ver.
	// Normally the previous version would be a different version, but
	// in forced upgrades, it can also be the save version string.
	// change from a non-nil meta needs to stop() the service first.
	// change to a non-nil meta must auto start() the service.
	change(from, to *drvapi.AppMeta) error

	// Send a soft signal to an app to start.
	start() error

	// Send a soft signal to an app to stop.
	stop() error
}

type appMaker interface {
	makeStub(name string) (*appStub, error)
}

type appStub struct {
	app
}
