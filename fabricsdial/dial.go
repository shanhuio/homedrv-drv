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

package fabricsdial

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"shanhu.io/g/sniproxy"
)

// NewWebSocketDialer creates a new WebSocket dialer from
// a http transport.
func NewWebSocketDialer(tr *http.Transport) *websocket.Dialer {
	return &websocket.Dialer{
		NetDialContext:  tr.DialContext,
		TLSClientConfig: tr.TLSClientConfig,
	}
}

// Dialer dials to a HomeDrive Fabrics service.
type Dialer struct {
	Router          sniproxy.Router
	WebSocketDialer *websocket.Dialer
	TunnelOptions   *sniproxy.Options
}

var defaultTunnelOptions = &sniproxy.Options{
	Siding:       true,
	DialWithAddr: true,
}

// Dial connects to a HomeDrive Fabrics service, and returns
// an SNI-proxy endpoint.
func (d *Dialer) Dial(ctx context.Context) (*sniproxy.Endpoint, error) {
	tunnOpts := d.TunnelOptions
	if tunnOpts == nil {
		tunnOpts = defaultTunnelOptions
	}
	opt := &sniproxy.DialOption{
		Path:          "/endpoint",
		TunnelOptions: tunnOpts,
		Dialer:        d.WebSocketDialer,
	}
	return sniproxy.Dial(ctx, d.Router, opt)
}
