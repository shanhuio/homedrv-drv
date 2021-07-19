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

package fabricsdial

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"shanhu.io/aries/creds"
	fabguest "shanhu.io/homedrv/fabricsguest"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/virgo/sniproxy"
)

// Dialer dials to a HomeDrive Fabrics service.
type Dialer struct {
	Host string
	User string

	Key     []byte
	KeyFile string

	// Guest is the callback function for receiving the guest domain. When
	// specified, the dialer dials in as guest, and callbacks with the guest
	// domain.
	Guest func(domain string) error

	TunnelOptions *sniproxy.Options

	Transport http.RoundTripper
}

func (d *Dialer) dialOption() (*sniproxy.DialOption, error) {
	if d.Guest != nil {
		c := &httputil.Client{
			Server: &url.URL{
				Scheme: "https",
				Host:   d.Host,
			},
			Transport: d.Transport,
		}
		info := &fabguest.Info{User: d.User}
		tok := new(fabguest.Token)
		if err := c.Call("/guest-token", info, tok); err != nil {
			return nil, errcode.Annotate(err, "fetch token")
		}
		if err := d.Guest(tok.Domain); err != nil {
			return nil, err
		}
		return &sniproxy.DialOption{GuestToken: tok.Token}, nil
	}

	cep := &creds.Endpoint{
		Server: (&url.URL{
			Scheme: "https",
			Host:   d.Host,
		}).String(),
		User:        d.User,
		Key:         d.Key,
		PemFile:     d.KeyFile,
		Homeless:    true,
		NoTTY:       true,
		NoPermCheck: true,
	}
	if d.Transport != nil {
		cep.Transport = d.Transport
	}
	login, err := creds.NewLogin(cep)
	if err != nil {
		return nil, errcode.Annotate(err, "create login")
	}
	tunnOptions := d.TunnelOptions
	if tunnOptions == nil {
		tunnOptions = &sniproxy.Options{
			Siding:       true,
			DialWithAddr: true,
		}
	}
	return &sniproxy.DialOption{
		Login:         login,
		TunnelOptions: tunnOptions,
	}, nil
}

// Dial connects to a HomeDrive Fabrics service, and returns
// an SNI-proxy endpoint.
func (d *Dialer) Dial(ctx context.Context) (*sniproxy.Endpoint, error) {
	opt, err := d.dialOption()
	if err != nil {
		return nil, err
	}

	proxyURL := &url.URL{
		Scheme: "wss",
		Host:   d.Host,
		Path:   "/endpoint",
	}

	if d.Transport != nil {
		tr, ok := d.Transport.(*http.Transport)
		if !ok {
			return nil, errors.New("transport is not an http transport")
		}
		opt.Dialer = &websocket.Dialer{
			NetDialContext:  tr.DialContext,
			TLSClientConfig: tr.TLSClientConfig,
		}
	}
	return sniproxy.Dial(ctx, proxyURL, opt)
}
