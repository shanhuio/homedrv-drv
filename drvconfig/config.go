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

package drvconfig

// Config is the configuration of jarvis. These configurations are critical for
// initializing an endpoint and are largely immutable.
type Config struct {
	Name string // Name of the endpoint.

	// Server address, default https://www.homedrive.io
	Server string `json:",omitempty"`

	// Pin to a particular build.
	Build string `json:",omitempty"`

	// Subscribe to which release channel.
	Channel string `json:",omitempty"`

	// Naming conventions. When this is null, using legacy naming.
	Naming *Naming `json:",omitempty"`

	// Identity PEM key file.
	IdentityPem string `json:",omitempty"`

	// Path to docker unix domain socket.
	DockerSock string `json:",omitempty"`

	// Path to system docker unix domain socket.
	SystemDockerSock string `json:",omitempty"`

	// Running outside a docker. Useful for testing.
	Bare bool `json:",omitempty"`

	// HTTPPort provides alternative http port for doorway container to
	// listen on. If it is negative, then doorway will not listen on
	// HTTP.
	HTTPPort int `json:",omitempty"`

	// HTTPSPort provides the alternative https port for doorway
	// container to listen on. If it is negative, then doorway will
	// not listen on HTTPS.
	HTTPSPort int `json:",omitempty"`

	// When this is true, do not bind ports when port number is 0 and it is not
	// managing the OS.
	AutoAvoidPortBinding bool `json:",omitempty"`

	// Instead of reading the endpoint init config from the server,
	// read from this file.
	EndpointConfigFile string `json:",omitempty"`

	// Dev mode device.
	Dev bool `json:",omitempty"`
}
