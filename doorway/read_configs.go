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
	"os"

	"golang.org/x/crypto/acme/autocert"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/jsonx"
	"shanhu.io/misc/osutil"
)

func readHostMap(p string) (map[string]string, error) {
	m := make(map[string]string)
	if err := jsonx.ReadFile(p, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func serverConfigFromHome(h *osutil.Home) (*ServerConfig, error) {
	hostMap, err := readHostMap(h.FilePath("etc/host-map.jsonx"))
	if err != nil {
		return nil, errcode.Annotate(err, "read host map")
	}
	certCacheDir := h.FilePath("var/autocert")
	dirExists, err := osutil.IsDir(certCacheDir)
	if err != nil {
		return nil, errcode.Annotate(err, "check cert cache dir")
	}
	if !dirExists {
		if err := os.Mkdir(certCacheDir, 0700); err != nil {
			return nil, errcode.Annotate(err, "make cert cache dir")
		}
	}
	return &ServerConfig{
		HostMap:       hostMap,
		AutoCertCache: autocert.DirCache(certCacheDir),
	}, nil
}

func httpServerConfigFromHome(h *osutil.Home) (*HTTPServerConfig, error) {
	config := new(HTTPServerConfig)
	p := h.FilePath("etc/http.jsonx")
	if err := jsonx.ReadFile(p, config); err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	return config, nil
}

func fabricsConfigFromHome(h *osutil.Home) (*FabricsConfig, error) {
	c := new(FabricsConfig)
	p := h.FilePath("etc/fabrics.jsonx")
	if err := jsonx.ReadFile(p, c); err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	return c, nil
}

// ConfigFromHome reads Config from the given directories.
func ConfigFromHome(homeDir string) (*Config, error) {
	h, err := osutil.NewHome(homeDir)
	if err != nil {
		return nil, errcode.Annotate(err, "make home")
	}

	c := new(Config)

	serverConfig, err := serverConfigFromHome(h)
	if err != nil {
		return nil, errcode.Annotate(err, "build server config")
	}
	c.Server = serverConfig

	httpConfig, err := httpServerConfigFromHome(h)
	if err != nil {
		return nil, errcode.Annotate(err, "read http server config")
	}
	c.HTTPServer = httpConfig

	fabConfig, err := fabricsConfigFromHome(h)
	if err != nil {
		return nil, errcode.Annotate(err, "read fabrics config")
	}

	if fabConfig.User != "" {
		c.Fabrics = fabConfig

		pemPath := h.FilePath("var/fabrics.pem")
		id, err := newFileIdentity(pemPath)
		if err != nil {
			return nil, errcode.Annotate(err, "read fabrics identity pem")
		}
		c.FabricsIdentity = id
	}

	return c, nil
}
