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
	"net/url"
	"path"
	"time"

	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/sqlx"
	"shanhu.io/virgo/dock"
)

type postgres struct {
	core appCore
}

func newPostgres(c appCore) *postgres {
	return &postgres{core: c}
}

func (p *postgres) cont() *dock.Cont {
	d := p.core.Docker()
	return dock.NewCont(d, appCont(p.core, namePostgres))
}

func (p *postgres) createCont(image, pwd string) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}
	if pwd == "" {
		return nil, errcode.InvalidArgf("database root password empty")
	}

	d := p.core.Docker()
	labels := drvcfg.NewNameLabel(namePostgres)
	volName := appVol(p.core, namePostgres)
	if _, err := dock.CreateVolumeIfNotExist(
		d, volName, &dock.VolumeConfig{Labels: labels},
	); err != nil {
		return nil, errcode.Annotate(err, "create postgres volume")
	}

	name := appCont(p.core, namePostgres)

	config := &dock.ContConfig{
		Name:    name,
		Network: appNetwork(p.core),
		Env:     map[string]string{"POSTGRES_PASSWORD": pwd},
		Mounts: []*dock.ContMount{{
			Type: dock.MountVolume,
			Host: volName,
			Cont: "/var/lib/postgresql/data",
		}},
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Labels:        labels,
	}
	return dock.CreateCont(d, image, config)
}

func (p *postgres) install(image string) error {
	pwd, err := p.password()
	if err != nil {
		return errcode.Annotate(err, "read password")
	}
	cont, err := p.createCont(image, pwd)
	if err != nil {
		return errcode.Annotate(err, "create container")
	}
	if err := cont.Start(); err != nil {
		return errcode.Annotate(err, "start postgres")
	}
	if err := p.startWait(); err != nil {
		return errcode.Annotate(err, "wait for db up")
	}
	return nil
}

func (p *postgres) update(image string) error {
	if image == "" {
		return errcode.InvalidArgf("postgres image empty")
	}
	contName := appCont(p.core, namePostgres)
	d := p.core.Docker()
	if err := dropContIfDifferent(d, contName, image); err != nil {
		if err == errSameImage {
			return nil
		}
		return err
	}
	log.Println("update postgres")
	return p.install(image)
}

func (p *postgres) open(user, pwd, db string) (*sqlx.DB, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pwd),
		Host:   appCont(p.core, namePostgres),
		Path:   path.Join("/", db),
	}
	q := make(url.Values)
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return sqlx.OpenPsql(u.String())
}

func (p *postgres) password() (string, error) {
	return readPasswordOrSetRandom(p.core.Settings(), keyPostgresPass)
}

func (p *postgres) openAdmin() (*sqlx.DB, error) {
	password, err := p.password()
	if err != nil {
		return nil, errcode.Annotate(err, "read password")
	}
	return p.open("postgres", password, "")
}

func (p *postgres) startWait() error {
	db, err := p.openAdmin()
	if err != nil {
		return errcode.Annotate(err, "open db")
	}
	defer db.Close()
	return waitDB(db, 5*time.Minute)
}

func (p *postgres) createDB(name, pwd string) error {
	db, err := p.openAdmin()
	if err != nil {
		return errcode.Annotate(err, "open db")
	}
	defer db.Close()
	return createDB(db, name, pwd)
}

func (p *postgres) dropDB(name string) error {
	db, err := p.openAdmin()
	if err != nil {
		return errcode.Annotate(err, "open db")
	}
	defer db.Close()
	return dropDB(db, name)
}

func (p *postgres) change(from, to *drvapi.AppMeta) error {
	if from != nil {
		if err := p.cont().Drop(); err != nil {
			return errcode.Annotate(err, "drop old postgres container")
		}
	}
	if to == nil {
		vol := appVol(p.core, namePostgres)
		if err := dock.RemoveVolume(p.core.Docker(), vol); err != nil {
			return errcode.Annotate(err, "remove volume")
		}
		return nil
	}

	pwd, err := p.password()
	if err != nil {
		return errcode.Annotate(err, "read password")
	}
	// TODO(h8liu): implement proper postgresql upgrade.
	cont, err := p.createCont(appImage(to), pwd)
	if err != nil {
		return errcode.Annotate(err, "create postgres container")
	}
	if err := cont.Start(); err != nil {
		return err
	}
	if err := p.startWait(); err != nil {
		return errcode.Annotate(err, "wait for db to start")
	}
	return nil
}

func (p *postgres) start() error { return p.cont().Start() }
func (p *postgres) stop() error  { return p.cont().Stop() }
