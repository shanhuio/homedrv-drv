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

// RegisterRequest is the request for setting up an endpoint's public key using
// a passcode.
type RegisterRequest struct {
	Name       string
	PassCode   string
	ControlKey string `json:",omitempty"`
}

// RegisterDoorwayRequest is the request for registering a doorway fabrics
// connection.
type RegisterDoorwayRequest struct {
	PublicKey string `json:",omitempty"`
}

// RegisterTunnelRequest is the request for registering a hometunn fabrics
// connection.
type RegisterTunnelRequest struct {
	Name      string
	PassCode  string
	PublicKey string
}
