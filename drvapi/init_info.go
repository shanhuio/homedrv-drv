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

// InitInfo contains information of the init procedure.
type InitInfo struct {
	Time              int64  `json:",omitempty"`
	TimeSec           int64  `json:",omitempty"`
	JarvisPassword    string `json:",omitempty"`
	NextcloudPassword string `json:",omitempty"`
}

// InitDoneRequest is the request to set the init info of an endpoint.
type InitDoneRequest struct {
	Name string
	Info *InitInfo
}

// InitDoneResponse is the response of setting the init info of an endpoint.
type InitDoneResponse struct{}
