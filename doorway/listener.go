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

	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/counting"
)

type localListenConfig struct {
	// listener is the listener to listen on.
	listener net.Listener

	// addr is the TCP address to listen on. This is used when listener is nil.
	addr string

	// counters track the number of bytes communicated over https listener.
	counters *counting.ConnCounters
}

type listenConfig struct {
	// local listens on local TCP address or a specific provided listener.
	local *localListenConfig

	// fabrics is for listening on a fabrics connection.
	fabrics *fabricsConfig
}

func listenLocal(c *localListenConfig) (*tagListener, error) {
	if c.listener != nil {
		return newTagListener(c.listener, tagTCP), nil
	}
	if c.addr == "" {
		return nil, errcode.InvalidArgf("listen address missing")
	}
	tcp, err := net.Listen("tcp", c.addr)
	if err != nil {
		return nil, errcode.Annotate(err, "listen local")
	}
	lis := counting.WrapListener(tcp, c.counters)
	return newTagListener(lis, tagTCP), nil
}

func listen(ctx C, c *listenConfig) (tagConnListener, error) {
	if c.local == nil && c.fabrics == nil {
		return nil, errcode.InvalidArgf(
			"must listen either at local or via fabrics",
		)
	}

	if c.fabrics == nil { // no fabrics, just listen local
		lis, err := listenLocal(c.local)
		if err != nil {
			return nil, err
		}
		return lis, nil
	}

	fab := newFabricsClient(c.fabrics)
	if c.local == nil {
		lis, err := listenFabrics(ctx, fab)
		if err != nil {
			return nil, err
		}
		return lis, nil
	}

	// Dual listener.
	local, err := listenLocal(c.local)
	if err != nil {
		return nil, err
	}
	fabLis, err := listenFabrics(ctx, fab)
	if err != nil {
		local.Close()
		return nil, errcode.Annotate(err, "listen fabrics")
	}
	return newTunnelListener(local, fabLis), nil
}
