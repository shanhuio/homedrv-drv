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

package doorway

import (
	"net"
)

func v4ZeroAddr(port int) *net.TCPAddr {
	return &net.TCPAddr{
		IP: net.IPv4zero, Port: port,
	}
}

func lisAddr(lis net.Listener) string {
	return lis.Addr().String()
}

var ipv4localhost = net.IPv4(127, 0, 0, 1)

func localV4Addr(port int) *net.TCPAddr {
	return &net.TCPAddr{
		IP:   ipv4localhost,
		Port: port,
	}
}

func allIfaceAddr(port int) *net.TCPAddr {
	return &net.TCPAddr{Port: port}
}
