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
	"shanhu.io/g/dock"
	"shanhu.io/g/settings"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
)

// Core provides the core interface to run an application.
type Core interface {
	// App gets an application by name.
	App(name string) (App, error)

	// Docker gets the client to the application docker.
	Docker() *dock.Client

	// Settings gets the settings table.
	Settings() settings.Settings

	// Naming gets the naming convention of the drive. We might want to
	// migrate the legacy stuff and deprecate this some day.
	Naming() *drvcfg.Naming

	// Domains gets the stub that manages application domain routings.
	Domains() Domains
}

// Cont returns the container name of an app.
func Cont(c Core, app string) string { return drvcfg.Name(c.Naming(), app) }

// Network returns the network name.
func Network(c Core) string { return drvcfg.Network(c.Naming()) }

// Vol returns the volume name of an app.
func Vol(c Core, app string) string { return drvcfg.Name(c.Naming(), app) }
