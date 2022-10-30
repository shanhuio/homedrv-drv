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

	"shanhu.io/aries/creds"
	"shanhu.io/misc/errcode"
)

// SimpleRouter provides a simple endpoint based router. It directly contacts
// the fabrics node for a token.
type SimpleRouter struct {
	Host    string // Host to route to.
	User    string
	Key     []byte
	KeyFile string

	Transport http.RoundTripper
}

// Route returns the host and the token.
func (r *SimpleRouter) Route(ctx context.Context) (string, string, error) {
	host := r.Host
	ep := &creds.Endpoint{
		Server:   &url.URL{Scheme: "https", Host: host},
		User:     r.User,
		Key:      r.Key,
		PemFile:  r.KeyFile,
		Homeless: true,
		NoTTY:    true,
	}
	if r.Transport != nil {
		ep.Transport = r.Transport
	}
	login, err := creds.NewLogin(ep)
	if err != nil {
		return "", "", errcode.Annotate(err, "create login")
	}
	token, err := login.Token()
	if err != nil {
		return "", "", errcode.Annotate(err, "login")
	}
	return host, token, nil
}
