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

package homeapp

import (
	"shanhu.io/homedrv/drv/drvapi"
)

// App is a generic application object that manages the lifecycle
// if an application running on a HomeDrive.
type App interface {
	// Called when the version is changed from a non-empty string to ver.
	// Normally the previous version would be a different version, but
	// in forced upgrades, it can also be the save version string.
	// change from a non-nil meta needs to stop() the service first.
	// change to a non-nil meta must auto start() the service.
	Change(from, to *drvapi.AppMeta) error

	// Send a soft signal to an app to start.
	Start() error

	// Send a soft signal to an app to stop.
	Stop() error
}
