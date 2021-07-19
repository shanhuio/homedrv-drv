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
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"golang.org/x/crypto/acme/autocert"
	"shanhu.io/aries"
	"shanhu.io/misc/errcode"
)

// ServerConfig is the config for serving the reverse proxy
// server.
type ServerConfig struct {
	HostMap       map[string]string
	AutoCertCache autocert.Cache
}

type server struct {
	home          *aries.Router
	hostMap       hostMap
	proxy         *httputil.ReverseProxy
	autoCertCache autocert.Cache
}

func newServer(config *ServerConfig) *server {
	s := &server{
		hostMap:       newMemHostMap(config.HostMap),
		autoCertCache: config.AutoCertCache,
	}

	home := aries.NewRouter()
	home.Index(aries.StringFunc("hi"))
	home.File("health", aries.StringFunc("ok"))
	s.home = home

	s.proxy = &httputil.ReverseProxy{
		Director:       s.director,
		ModifyResponse: setStrictTransportSecurity,
	}
	return s
}

func (s *server) Serve(c *aries.C) error {
	host := strings.TrimSuffix(c.Req.Host, ".")

	entry := s.hostMap.mapHost(host)
	if entry == nil {
		return aries.NotFound
	}

	switch entry.typ {
	default:
		return aries.NotFound
	case hostHome:
		return s.serveHome(c)
	case hostRedirect:
		u := *c.Req.URL
		u.Host = entry.host
		c.Redirect(u.String())
		return nil
	case hostProxy:
		s.proxy.ServeHTTP(c.Resp, c.Req)
		return nil
	}
}

func (s *server) serveHome(c *aries.C) error {
	return s.home.Serve(c)
}

func (s *server) director(req *http.Request) {
	// swap the scheme to http
	req.Header.Set("X-Forwarded-Proto", "https")

	host := req.Host

	mapped := hostMapToProxy(s.hostMap, host)
	if mapped == "" {
		if host == "" {
			log.Println("empty host")
		} else {
			log.Printf("unexpected host: %q", host)
		}

		req.URL = sinkURL
		return
	}

	forwardToHTTP(req, mapped)
}

func setStrictTransportSecurity(resp *http.Response) error {
	resp.Header.Set(
		"Strict-Transport-Security",
		"max-age=15552000; includeSubDomains",
	)
	return nil
}

// hostPolicy determines which hosts are whitelisted for autocert.
func (s *server) hostPolicy(_ context.Context, host string) error {
	if !hostMapHas(s.hostMap, host) {
		return errcode.NotFoundf("%q not in whitelist", host)
	}
	return nil
}

// HostMapRequest is the request for host map editing.
type HostMapRequest struct {
	From string
	To   string
}

func (s *server) autoTLSConfig() *tls.Config {
	autoCert := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: s.hostPolicy,
		Cache:      s.autoCertCache,
	}
	return autoCert.TLSConfig()
}
