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
	"context"
	"log"
	"net"
	"net/http"

	"shanhu.io/aries/https/httpstest"
	fabdial "shanhu.io/homedrv/fabricsdial"
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
	dialerFunc func(ctx context.Context) (*fabdial.Dialer, error)

	*FabricsConfig
	identity Identity

	dialTransport     http.RoundTripper
	registerTransport http.RoundTripper

	// counters track the number of bytes communicated over the tunnel.
	counters *counting.ConnCounters
}

type fabricsClient struct {
	config *fabricsConfig // Dialer from configuration.
}

func newFabricsClient(config *fabricsConfig) *fabricsClient {
	return &fabricsClient{config: config}
}

func (f *fabricsClient) dialer(ctx C) (*fabdial.Dialer, error) {
	config := f.config
	if f := config.dialerFunc; f != nil {
		return f(ctx)
	}

	key, err := f.config.identity.Load(ctx)
	if err != nil {
		return nil, errcode.Annotate(err, "read fabrics key")
	}

	dialer := &fabdial.Dialer{
		Host:      config.host(),
		User:      config.User,
		Key:       key,
		Transport: f.config.dialTransport,
	}

	if config.InsecurelyDialTo != "" {
		dialer.Transport = httpstest.InsecureSink(config.InsecurelyDialTo)
	}
	return dialer, nil
}

func listenFabrics(ctx C, c *fabricsClient) (net.Listener, error) {
	d, err := c.dialer(ctx)
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
	return counting.WrapListener(lis, c.config.counters), nil
}
