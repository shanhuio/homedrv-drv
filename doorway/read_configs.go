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
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/jsonx"
)

func readHostMap(etcDir string) (map[string]string, error) {
	m := make(map[string]string)
	p := filepath.Join(etcDir, "host-map.jsonx")
	if err := jsonx.ReadFile(p, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func serverConfigFromDirs(etcDir, varDir string) (*ServerConfig, error) {
	hostMap, err := readHostMap(etcDir)
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		HostMap:       hostMap,
		AutoCertCache: autocert.DirCache(filepath.Join(varDir, "autocert")),
	}, nil
}

func httpServerConfigFromDir(etcDir string) (*HTTPServerConfig, error) {
	config := new(HTTPServerConfig)
	p := filepath.Join(etcDir, "http.jsonx")
	if err := jsonx.ReadFile(p, config); err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	return config, nil
}

func readFabricsConfig(etcDir string) (*FabricsConfig, error) {
	c := new(FabricsConfig)
	p := filepath.Join(etcDir, "fabrics.jsonx")
	if err := jsonx.ReadFile(p, c); err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	return c, nil
}

// ConfigFromDirs reads Config from the given directories.
func ConfigFromDirs(etcDir, varDir string) (*Config, error) {
	c := new(Config)

	serverConfig, err := serverConfigFromDirs(etcDir, varDir)
	if err != nil {
		return nil, errcode.Annotate(err, "build server config")
	}
	c.Server = serverConfig

	httpConfig, err := httpServerConfigFromDir(etcDir)
	if err != nil {
		return nil, errcode.Annotate(err, "read http server config")
	}
	c.HTTPServer = httpConfig

	fabConfig, err := readFabricsConfig(etcDir)
	if err != nil {
		return nil, errcode.Annotate(err, "read fabrics config")
	}

	if fabConfig.User != "" {
		c.Fabrics = fabConfig

		pemPath := filepath.Join(varDir, "fabrics.pem")
		id, err := newFileIdentity(pemPath)
		if err != nil {
			return nil, errcode.Annotate(err, "read fabrics identity pem")
		}
		c.FabricsIdentity = id
	}

	return c, nil
}
