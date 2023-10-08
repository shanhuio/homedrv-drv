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

package jarvis

import (
	"log"
	"sort"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/g/rsautil"
	"shanhu.io/g/tarutil"
	doorwaypkg "shanhu.io/homedrv/drv/doorway"
	"shanhu.io/homedrv/drv/drvapi"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
)

const (
	doorwayEtcDir  = "/opt/app/etc"
	doorwayVarDir  = "/opt/app/var"
	doorwayUser    = 3000
	doorwayHostMap = "host-map.jsonx"
)

type doorwayConfig struct {
	domain        string
	fabricsServer string
	noFabrics     bool
}

type doorway struct {
	*drive
	config *doorwayConfig
}

func newDoorway(d *drive, config *doorwayConfig) *doorway {
	return &doorway{
		drive:  d,
		config: config,
	}
}

func (d *doorway) tarMeta(mode int64) *tarutil.Meta {
	return &tarutil.Meta{
		Mode:    mode,
		UserID:  doorwayUser,
		GroupID: doorwayUser,
	}
}

func (d *doorway) generateFabricsKey(t *tarutil.Stream) error {
	pri, pub, err := rsautil.GenerateKey(nil, 0)
	if err != nil {
		return err
	}
	t.AddBytes("fabrics.pem", d.tarMeta(0600), pri)
	t.AddBytes("fabrics.pub", d.tarMeta(0644), pub)

	if d.hasServer() {
		client, err := d.dialServer()
		if err != nil {
			return errcode.Annotate(err, "dial server")
		}
		const p = "/pubapi/endpoint/register:doorway"
		req := &drvapi.RegisterDoorwayRequest{PublicKey: string(pub)}
		if err := client.Call(p, req, nil); err != nil {
			return errcode.Annotate(err, "register to fabrics")
		}
	}
	return nil
}

func (d *doorway) coreAddr() string {
	// TODO(h8liu): read the port from start config
	return d.core() + ":3377"
}

func (d *doorway) hostMap() (map[string]string, error) {
	m := make(map[string]string)

	subs, err := loadCustomSubs(d.settings)
	if err != nil {
		return nil, errcode.Annotate(err, "check custom subs")
	}
	for sub, dest := range subs {
		m[sub] = dest
	}

	apps, err := d.appDomains.list()
	if err != nil {
		return nil, errcode.Annotate(err, "list app domains")
	}
	for _, app := range apps {
		var domains []string
		for domain := range app.Map {
			domains = append(domains, domain)
		}
		sort.Strings(domains)
		for _, domain := range domains {
			m[domain] = app.Map[domain].Dest
		}
	}

	if !d.drive.config.External {
		if main := d.config.domain; main != "" {
			m[main] = d.coreAddr()
		}
	}

	return m, nil
}

func (d *doorway) etcFiles() (*tarutil.Stream, error) {
	s := tarutil.NewStream()

	if !d.config.noFabrics {
		fc := &doorwaypkg.FabricsConfig{
			User: d.name,
			Host: d.config.fabricsServer,
		}
		if err := addJSONXToTarStream(
			s, "fabrics.jsonx", d.tarMeta(0600), fc,
		); err != nil {
			return nil, errcode.Annotate(err, "prepare fabrics config")
		}
	}
	m, err := d.hostMap()
	if err != nil {
		return nil, errcode.Annotate(err, "make host map")
	}

	if err := addJSONXToTarStream(
		s, doorwayHostMap, d.tarMeta(0600), m,
	); err != nil {
		return nil, errcode.Annotate(err, "prepare host map")
	}
	return s, nil
}

func (d *doorway) initVarFiles() (*tarutil.Stream, error) {
	s := tarutil.NewStream()
	if !d.config.noFabrics {
		if err := d.generateFabricsKey(s); err != nil {
			return nil, errcode.Annotate(err, "generate fabrics key")
		}
	}
	return s, nil
}

func (d *doorway) install(image string) error {
	if image == "" {
		image = d.image(nameDoorway)
	}
	varFiles, err := d.initVarFiles()
	if err != nil {
		return errcode.Annotate(err, "init var files")
	}
	return d.start(image, varFiles)
}

func (d *doorway) update(image string) error {
	c := dock.NewCont(d.dock, d.cont(nameDoorway))
	if _, err := c.Inspect(); err != nil {
		if errcode.IsNotFound(err) {
			log.Printf("doorway not found. try to reinstall.")
			return d.install(image)
		}
		return errcode.Annotate(err, "inspect doorway")
	}

	if err := c.Stop(); err != nil {
		return errcode.Annotate(err, "stop doorway")
	}
	if err := c.Drop(); err != nil {
		return errcode.Annotate(err, "drop doorway")
	}
	return d.start(image, nil)
}

func shouldBindPort0(d *drive) bool {
	return !d.config.AutoAvoidPortBinding || d.hasSys()
}

func (d *doorway) start(
	image string, varFiles *tarutil.Stream,
) error {
	etcFiles, err := d.etcFiles()
	if err != nil {
		return errcode.Annotate(err, "build etc files")
	}

	labels := drvcfg.NewNameLabel(nameDoorway)

	volName := d.vol(nameDoorway)
	if _, err := dock.CreateVolumeIfNotExist(
		d.dock, volName, &dock.VolumeConfig{Labels: labels},
	); err != nil {
		return errcode.Annotate(err, "create doorway volume")
	}

	var portBinds []*dock.PortBind
	// TODO(h8liu): read from settings rather than drive config.
	// makes the port binding configurable.
	drvConfig := d.drive.config
	if port := drvConfig.HTTPPort; port >= 0 {
		if port == 0 && shouldBindPort0(d.drive) {
			port = 80
		}
		if port > 0 {
			portBinds = append(portBinds, &dock.PortBind{
				HostPort: port, ContPort: 8080,
			})
		}
	}
	if port := drvConfig.HTTPSPort; port >= 0 {
		if port == 0 && shouldBindPort0(d.drive) {
			port = 443
		}
		if port > 0 {
			portBinds = append(portBinds, &dock.PortBind{
				HostPort: port, ContPort: 8443,
			})
		}
	}

	config := &dock.ContConfig{
		Name:        d.cont(nameDoorway),
		Network:     d.network(),
		AutoRestart: true,
		Mounts: []*dock.ContMount{{
			Type: dock.MountVolume,
			Host: volName,
			Cont: doorwayVarDir,
		}},
		TCPBinds:      portBinds,
		Labels:        labels,
		JSONLogConfig: dock.LimitedJSONLog(),
	}
	cont, err := dock.CreateCont(d.dock, image, config)
	if err != nil {
		return errcode.Annotate(err, "create doorway container")
	}

	if err := dock.CopyInTarStream(
		cont, etcFiles, doorwayEtcDir,
	); err != nil {
		cont.Drop()
		return errcode.Annotate(err, "copy in config files")
	}

	if varFiles != nil {
		if err := dock.CopyInTarStream(
			cont, varFiles, doorwayVarDir,
		); err != nil {
			cont.Drop()
			return errcode.Annotate(err, "copy in init var files")
		}
	}

	if err := cont.Start(); err != nil {
		return errcode.Annotate(err, "start doorway container")
	}
	return nil
}

// pingDomains pings all the domains 3 times and logs any error.  This
// mitigates the letsencrypt certificate timestamp issue, by trying to
// trigger the certificate issuing before the first visit from a user.
func (d *doorway) pingDomains() {
	m, err := d.hostMap()
	if err != nil {
		log.Println("fail to get host map: ", err)
		return
	}

	var domains []string
	for d := range m {
		domains = append(domains, d)
	}
	pingDomains(domains)
}
