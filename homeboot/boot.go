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

package homeboot

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"shanhu.io/aries/creds"
	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/flagutil"
	"shanhu.io/misc/httputil"
	"shanhu.io/misc/jsonx"
	"shanhu.io/misc/rsautil"
	"shanhu.io/misc/tarutil"
	"shanhu.io/virgo/dock"
)

// BootConfig is a json marshallable file that is saved on
// the file system, often as /opt/homedrv/boot.jsonx
// It specifies the flags used by homeinstall.
type BootConfig struct {
	Drive        *drvcfg.Config
	Code         string
	Download     bool `json:",omitempty"`
	LegacyNaming bool
}

func stableChannel() string {
	if runtime.GOARCH == "amd64" {
		return "stable"
	}
	// Architecture other than amd64 should use their corresponding
	// release channels.
	return "stable-" + runtime.GOARCH
}

func (c *BootConfig) declareFlags(flags *flagutil.FlagSet) {
	drv := c.Drive
	flags.StringVar(
		&drv.Server, "server", defaultServer, "server to register",
	)
	flags.StringVar(&drv.Name, "name", "", "endpoint name")
	flags.StringVar(&c.Code, "code", "", "registration one time passcode")
	flags.StringVar(&drv.Build, "build", "", "build to install")
	flags.StringVar(
		&drv.Channel, "channel", stableChannel(),
		"release channel to subscribe",
	)
	flags.StringVar(
		&drv.DockerSock, "docker", "", "docker unix domain socket",
	)
	flags.BoolVar(&c.Download, "download", true, "download docker image")
	flags.BoolVar(
		&c.LegacyNaming, "legacy_naming", false,
		"uses legacy naming, when used, -network is ignored",
	)
	flags.IntVar(
		&drv.HTTPPort, "http_port", 0,
		"http port to bind, "+
			"0 means 80 when managing OS or not auto_avoid_port_binding, "+
			"-1 means no binding",
	)
	flags.IntVar(
		&drv.HTTPSPort, "https_port", 0,
		"https port to bind, "+
			"0 means 443 when managing OS or not auto_avoid_port_binding, "+
			"-1 means no binding",
	)
	flags.BoolVar(
		&drv.AutoAvoidPortBinding, "auto_avoid_port_binding", true,
		"avoid binding ports when the port is 0 and not managing the OS",
	)
	flags.BoolVar(&drv.Dev, "dev", false, "enable dev mode")
}

func (c *BootConfig) fixLegacyNaming() {
	if c.LegacyNaming {
		c.Drive.Naming = nil
	}
}

func newBootConfig() *BootConfig {
	return &BootConfig{
		Drive: &drvcfg.Config{Naming: &drvcfg.Naming{}},
	}
}

type boot struct {
	*BootConfig
}

func newBoot(config *BootConfig) *boot {
	return &boot{BootConfig: config}
}

func writeFile(f string, bs []byte, mode os.FileMode) error {
	dir := filepath.Dir(f)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return errcode.Annotate(err, "make key directory")
		}
	}
	return ioutil.WriteFile(f, bs, mode)
}

func (b *boot) downloadCore(
	dock *dock.Client, config *drvcfg.Config, pem []byte,
) (string, error) {
	drv := b.Drive
	user := "~" + drv.Name
	ep := creds.NewRobot(user, drv.Server, "", nil)
	ep.Key = pem
	c, err := creds.DialEndpoint(ep)
	if err != nil {
		return "", err
	}

	d := NewOfficialDownloader(c, dock)
	rel, err := d.DownloadRelease(&InstallConfig{
		Build:    drv.Build,
		Channel:  drv.Channel,
		CoreOnly: true,
		Naming:   config.Naming,
	})
	if err != nil {
		return "", errcode.Annotate(err, "load core image")
	}

	return rel.Jarvis, nil
}

func (b *boot) saveDriveConfig(
	files *tarutil.Stream, c *drvcfg.Config,
) error {
	bs, err := jsonx.Marshal(c)
	if err != nil {
		return errcode.Annotate(err, "marshal core config")
	}
	if err != nil {
		return errcode.Annotate(err, "generate core config")
	}
	files.AddBytes("config.jsonx", tarutil.ModeMeta(0644), bs)
	return nil
}

func registerEndpoint(server *url.URL, name, code string, pub []byte) error {
	client := &httputil.Client{Server: server}
	const p = "/pubapi/endpoint/register"
	req := &drvapi.RegisterRequest{
		Name:       name,
		PassCode:   code,
		ControlKey: strings.TrimSpace(string(pub)),
	}
	return client.Call(p, req, nil)
}

func (b *boot) run() error {
	drv := b.Drive
	client := dock.NewUnixClient(drv.DockerSock)

	serverURL, err := url.Parse(drv.Server)
	if err != nil {
		return errcode.Annotate(err, "invalid server url")
	}

	pri, pub, err := rsautil.GenerateKey(nil, 0)
	if err != nil {
		return errcode.Annotate(err, "generate identity")
	}

	files := tarutil.NewStream()
	files.AddBytes("jarvis.pem", tarutil.ModeMeta(0600), pri)
	files.AddBytes("jarvis.pub", tarutil.ModeMeta(0644), pub)

	if err := b.saveDriveConfig(files, drv); err != nil {
		return err
	}

	if err := registerEndpoint(
		serverURL, drv.Name, b.Code, pub,
	); err != nil {
		return errcode.Annotate(err, "register endpoint")
	}
	log.Println("endpoint registered")

	var image string
	if b.Download {
		dl, err := b.downloadCore(client, drv, pri)
		if err != nil {
			return errcode.Annotate(err, "download core docker")
		}
		image = dl
		log.Println("HomeDrive core downloaded")
	}

	hasSysDock := true
	if err := CheckSystemDock(); err != nil {
		if !errcode.IsNotFound(err) {
			return errcode.Annotate(err, "check system docker socket")
		}
		hasSysDock = false
	}

	config := &CoreConfig{
		Drive:       drv,
		Image:       image,
		Files:       files,
		BindSysDock: hasSysDock,
	}
	id, err := b.startCore(client, config)
	if err != nil {
		return err
	}

	log.Printf("HomeDrive core started: %s", id)

	core := drvcfg.Core(drv.Naming)
	logsCmd := fmt.Sprintf("docker logs --follow %s", core)
	log.Printf("To track the installation progress, run: \n  %s", logsCmd)

	return nil
}

func (b *boot) startCore(
	client *dock.Client, config *CoreConfig,
) (string, error) {
	network := drvcfg.Network(config.Drive.Naming)

	found, err := dock.HasNetwork(client, network)
	if err != nil {
		return "", errcode.Annotatef(err, "check network: %q", network)
	}
	if !found {
		log.Printf("creating network %q ...", network)
		if err := dock.CreateNetwork(client, network); err != nil {
			return "", errcode.Annotatef(
				err, "create network: %q", network,
			)
		}
	}
	return StartCore(client, config)
}
