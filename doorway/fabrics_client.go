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
	"log"
	"net"
	"net/http"

	"shanhu.io/aries/https/httpstest"
	fabdial "shanhu.io/homedrv/drv/fabricsdial"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/counting"
)

// FabricsConfig has the configuration for connecting HomeDrive Fabrics.
// This config is JSON marshallable.
type FabricsConfig struct {
	User string
	Host string `json:",omitempty"` // Default using fabrics.homedrive.io

	InsecurelyDialTo string `json:",omitempty"`
}

func (c *FabricsConfig) host() string {
	if c.Host == "" {
		return "fabrics.homedrive.io"
	}
	return c.Host
}

type fabricsConfig struct {
	// Explicit dialer creater. Will use this dialer instead of the User:Host
	// when this is explicitly specified.
	dialer *fabdial.Dialer

	*FabricsConfig
	identity Identity

	dialTransport http.RoundTripper

	// counters track the number of bytes communicated over the tunnel.
	counters *counting.ConnCounters
}

func makeFabricsDialer(ctx C, config *fabricsConfig) (
	*fabdial.Dialer, error,
) {
	if config.dialer != nil {
		return config.dialer, nil
	}

	key, err := config.identity.Load(ctx)
	if err != nil {
		return nil, errcode.Annotate(err, "read fabrics key")
	}

	dialer := &fabdial.Dialer{
		Host:      config.host(),
		User:      config.User,
		Key:       key,
		Transport: config.dialTransport,
	}

	if config.InsecurelyDialTo != "" {
		dialer.Transport = httpstest.InsecureSink(config.InsecurelyDialTo)
	}
	return dialer, nil
}

func listenFabrics(ctx C, config *fabricsConfig) (*tagListener, error) {
	d, err := makeFabricsDialer(ctx, config)
	if err != nil {
		return nil, err
	}
	lis, err := newReconnectListener(
		func() (net.Listener, error) {
			ep, err := d.Dial(ctx)
			if err != nil {
				return nil, errcode.Annotatef(err, "dial proxy")
			}
			return &fabricsListener{Endpoint: ep}, nil
		},
		func(err error) { log.Println("fabrics connection: ", err) },
	)
	if err != nil {
		return nil, errcode.Annotatef(err, "dial fabrics")
	}
	wrap := counting.WrapListener(lis, config.counters)
	return newTagListener(wrap, tagFabrics), nil
}
