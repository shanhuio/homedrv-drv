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

package drvapi

// EndpointConfig is the basic settings of an endpoint. This only affects the
// init time of an endpoint. After an endpoint is provisioned, a user might
// be able to change the configuration via jarvis' user interface.
type EndpointConfig struct {
	// Main domain. Will serve jarvis Web UI. Currently redirects to the first
	// Nextcloud domain. If missing, an endpoint's main domain is
	// <name>.homedrv.com
	MainDomain string `json:",omitempty"`

	// Apps is the list of apps to install on initialization.
	Apps []string `json:",omitempty"`

	// Nextcloud domains. Incoming traffic for these domains will be redirected
	// to nextcloud. If empty, will use nextcloud.<name>.homedrv.com.
	NextcloudDomains []string `json:",omitempty"`

	// Extra domains that will be routed to this endpoint. Using those needs
	// custom doorway host map settings.
	ExtraDomains []string `json:",omitempty"`

	// Fabrics server to connect to. Default using "fabrics.homedrive.io"
	FabricsServer string `json:",omitempty"`
}
