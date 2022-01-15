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
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"shanhu.io/aries"
	fabdial "shanhu.io/homedrv/fabricsdial"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/netutil"
	"shanhu.io/virgo/counting"
	"shanhu.io/virgo/sniproxy"
)

// Config is the config of a doorway.
type Config struct {
	// Server is the config for the http server.
	// This also includes the reverse proxy.
	Server *ServerConfig

	// TLSProxy is the configuration for the TLS proxy.
	TLSProxy *TLSProxyConfig

	// HTTPServer is the config for the http server part.
	HTTPServer *HTTPServerConfig

	// Local address to listen on.
	LocalAddr string

	Fabrics         *FabricsConfig // Config for dialing fabrics.
	FabricsIdentity Identity       // Identity for dialing fabrics.

	// Alternative fabrics dialer.
	FabricsDialer *fabdial.Dialer

	// TLSConfig is for the TLS config for serving the service via https.
	// If not specified, autocert from Letsencrypt will be used.
	TLSConfig *tls.Config

	// ListenDone is the callback function when listen is done.
	ListenDone func()
}

type internalConfig struct {
	server     *ServerConfig
	tlsProxy   *TLSProxyConfig
	listen     *listenConfig
	listenDone func()
	tlsConfig  *tls.Config
}

func makeInternalConfig(config *Config) *internalConfig {
	lisConfig := new(listenConfig)
	if config.LocalAddr != "" {
		lisConfig.local = &localListenConfig{
			addr:     config.LocalAddr,
			counters: counting.NewConnCounters(),
		}
	}
	if config.FabricsDialer != nil {
		lisConfig.fabrics = &fabricsConfig{
			dialer:   config.FabricsDialer,
			counters: counting.NewConnCounters(),
		}
	} else if config.Fabrics != nil {
		lisConfig.fabrics = &fabricsConfig{
			FabricsConfig: config.Fabrics,
			identity:      config.FabricsIdentity,
			counters:      counting.NewConnCounters(),
		}
	}

	return &internalConfig{
		server:     config.Server,
		tlsProxy:   config.TLSProxy,
		listen:     lisConfig,
		tlsConfig:  config.TLSConfig,
		listenDone: config.ListenDone,
	}
}

// Serve serves doorway with the given config.
func Serve(ctx C, config *Config) error {
	if config.HTTPServer != nil {
		http := newHTTPServer(config.HTTPServer)
		go runHTTPServer(http)
	}

	internal := makeInternalConfig(config)
	return serve(ctx, internal)
}

func serve(ctx C, config *internalConfig) error {
	server, err := newServer(config.server)
	if err != nil {
		return errcode.Annotate(err, "make server")
	}

	lis, err := listen(ctx, config.listen)
	if err != nil {
		return errcode.Annotate(err, "listen")
	}

	var httpsLis net.Listener = lis
	if config.tlsProxy != nil {
		httpsLis = newTLSProxy(lis, config.tlsProxy)
	}
	defer httpsLis.Close()

	if config.listenDone != nil {
		config.listenDone()
	}

	tlsConfig := config.tlsConfig
	if tlsConfig == nil {
		tlsConfig = server.autoTLSConfig()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Printf("starts https on %q", lisAddr(httpsLis))
	https := &http.Server{
		TLSConfig: tlsConfig,
		Handler:   aries.Serve(server),
	}
	go func() {
		<-ctx.Done()
		https.Close()
	}()

	keepAlive := netutil.WrapKeepAlive(httpsLis)
	if err := https.ServeTLS(keepAlive, "", ""); err != nil {
		if sniproxy.IsClosedConnError(err) {
			return nil
		}
		return err
	}
	return nil
}
