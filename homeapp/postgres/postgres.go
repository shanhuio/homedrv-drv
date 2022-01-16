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

package postgres

import (
	"log"
	"net/url"
	"path"
	"time"

	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/homedrv/homeapp"
	"shanhu.io/homedrv/homeapp/apputil"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/sqlx"
	"shanhu.io/virgo/dock"
)

// Name is the app's name.
const Name = "postgres"

// KeyPass is key to postgresql root password.
const KeyPass = "postgress.pass"

// Postgres is the postgresql database app.
type Postgres struct {
	core homeapp.Core
}

// New creates a new postgres app.
func New(c homeapp.Core) *Postgres {
	return &Postgres{core: c}
}

func (p *Postgres) cont() *dock.Cont {
	d := p.core.Docker()
	return dock.NewCont(d, homeapp.Cont(p.core, Name))
}

func (p *Postgres) createCont(image, pwd string) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}
	if pwd == "" {
		return nil, errcode.InvalidArgf("database root password empty")
	}

	d := p.core.Docker()
	labels := drvcfg.NewNameLabel(Name)
	volName := homeapp.Vol(p.core, Name)
	if _, err := dock.CreateVolumeIfNotExist(
		d, volName, &dock.VolumeConfig{Labels: labels},
	); err != nil {
		return nil, errcode.Annotate(err, "create postgres volume")
	}

	name := homeapp.Cont(p.core, Name)

	config := &dock.ContConfig{
		Name:    name,
		Network: homeapp.Network(p.core),
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

func (p *Postgres) install(image string) error {
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

func (p *Postgres) update(image string) error {
	if image == "" {
		return errcode.InvalidArgf("postgres image empty")
	}
	contName := homeapp.Cont(p.core, Name)
	d := p.core.Docker()
	if err := apputil.DropIfDifferent(d, contName, image); err != nil {
		if err == apputil.ErrSameImage {
			return nil
		}
		return err
	}
	log.Println("update postgres")
	return p.install(image)
}

func (p *Postgres) open(user, pwd, db string) (*sqlx.DB, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pwd),
		Host:   homeapp.Cont(p.core, Name),
		Path:   path.Join("/", db),
	}
	q := make(url.Values)
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return sqlx.OpenPsql(u.String())
}

func (p *Postgres) password() (string, error) {
	return apputil.ReadPasswordOrSetRandom(p.core.Settings(), KeyPass)
}

func (p *Postgres) openAdmin() (*sqlx.DB, error) {
	password, err := p.password()
	if err != nil {
		return nil, errcode.Annotate(err, "read password")
	}
	return p.open("postgres", password, "")
}

func (p *Postgres) startWait() error {
	db, err := p.openAdmin()
	if err != nil {
		return errcode.Annotate(err, "open db")
	}
	defer db.Close()
	return waitDB(db, 5*time.Minute)
}

// CreateDB creates a new database.
func (p *Postgres) CreateDB(name, pwd string) error {
	db, err := p.openAdmin()
	if err != nil {
		return errcode.Annotate(err, "open db")
	}
	defer db.Close()
	return createDB(db, name, pwd)
}

// DropDB drops a database.
func (p *Postgres) DropDB(name string) error {
	db, err := p.openAdmin()
	if err != nil {
		return errcode.Annotate(err, "open db")
	}
	defer db.Close()
	return dropDB(db, name)
}

// Change changes the version from one to another.
func (p *Postgres) Change(from, to *drvapi.AppMeta) error {
	if from != nil {
		if err := p.cont().Drop(); err != nil {
			return errcode.Annotate(err, "drop old postgres container")
		}
	}
	if to == nil {
		vol := homeapp.Vol(p.core, Name)
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
	cont, err := p.createCont(homeapp.Image(to), pwd)
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

// Start starts the app.
func (p *Postgres) Start() error { return p.cont().Start() }

// Stop stops the app.
func (p *Postgres) Stop() error { return p.cont().Stop() }
