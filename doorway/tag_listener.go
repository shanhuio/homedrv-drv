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

package doorway

import (
	"net"
)

const (
	tagTCP     = "TCP"
	tagFabrics = "fabrics"
)

type tagConnListener interface {
	net.Listener
	acceptTag() (*tagConn, error)
}

type tagListener struct {
	net.Listener
	tag string
}

func newTagListener(lis net.Listener, tag string) *tagListener {
	return &tagListener{
		Listener: lis,
		tag:      tag,
	}
}

func (l *tagListener) Accept() (net.Conn, error) {
	conn, err := l.acceptTag()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (l *tagListener) acceptTag() (*tagConn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &tagConn{Conn: conn, tag: l.tag}, nil
}

type tagConn struct {
	net.Conn
	tag string
}
