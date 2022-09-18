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

package fabricsdial

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"shanhu.io/aries/creds"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/sniproxy"
)

// Dialer dials to a HomeDrive Fabrics service.
type Dialer struct {
	HostTokenFunc func(ctx context.Context) (string, string, error)

	Host    string
	User    string
	Key     []byte
	KeyFile string

	TunnelOptions *sniproxy.Options

	Transport *http.Transport
}

func (d *Dialer) hostToken(ctx context.Context) (string, string, error) {
	if d.HostTokenFunc != nil {
		return d.HostTokenFunc(ctx)
	}

	host := d.Host
	cep := &creds.Endpoint{
		Server: (&url.URL{
			Scheme: "https",
			Host:   host,
		}).String(),
		User:     d.User,
		Key:      d.Key,
		PemFile:  d.KeyFile,
		Homeless: true,
		NoTTY:    true,
	}
	if d.Transport != nil {
		cep.Transport = d.Transport
	}
	login, err := creds.NewLogin(cep)
	if err != nil {
		return "", "", errcode.Annotate(err, "create login")
	}
	token, err := login.Token()
	if err != nil {
		return "", "", errcode.Annotate(err, "login")
	}
	return host, token, nil
}

func (d *Dialer) dialOption(tok string) (*sniproxy.DialOption, error) {
	tunnOptions := d.TunnelOptions
	if tunnOptions == nil {
		tunnOptions = &sniproxy.Options{
			Siding:       true,
			DialWithAddr: true,
		}
	}
	return &sniproxy.DialOption{
		Token:         tok,
		TunnelOptions: tunnOptions,
	}, nil
}

// Dial connects to a HomeDrive Fabrics service, and returns
// an SNI-proxy endpoint.
func (d *Dialer) Dial(ctx context.Context) (*sniproxy.Endpoint, error) {
	host, tok, err := d.hostToken(ctx)
	if err != nil {
		return nil, errcode.Annotate(err, "pick server")
	}

	opt, err := d.dialOption(tok)
	if err != nil {
		return nil, errcode.Annotate(err, "prepare options")
	}

	proxyURL := &url.URL{
		Scheme: "wss",
		Host:   host,
		Path:   "/endpoint",
	}

	if tr := d.Transport; tr != nil {
		opt.Dialer = &websocket.Dialer{
			NetDialContext:  tr.DialContext,
			TLSClientConfig: tr.TLSClientConfig,
		}
	}
	return sniproxy.Dial(ctx, proxyURL, opt)
}
