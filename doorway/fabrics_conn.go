// Copyright (C) 2023  Shanhu Tech Inc.
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

	"shanhu.io/g/sniproxy"
)

type fabricsAddr struct {
	addr string
}

func (*fabricsAddr) Network() string  { return "tcp" }
func (a *fabricsAddr) String() string { return "|" + a.addr }

type fabricsConn struct {
	net.Conn
}

func (c *fabricsConn) Addr() net.Addr {
	return &fabricsAddr{addr: c.Conn.RemoteAddr().String()}
}

type fabricsListener struct {
	*sniproxy.Endpoint
}

func (l *fabricsListener) Accept() (net.Conn, error) {
	conn, err := l.Endpoint.Accept()
	if err != nil {
		return nil, err
	}
	return &fabricsConn{Conn: conn}, nil
}
