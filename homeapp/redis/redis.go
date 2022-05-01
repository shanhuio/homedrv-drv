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

package redis

import (
	"fmt"

	"shanhu.io/homedrv/drv/drvapi"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
	"shanhu.io/homedrv/drv/homeapp"
	"shanhu.io/homedrv/drv/homeapp/apputil"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/tarutil"
	"shanhu.io/virgo/dock"
)

// Name is the name of the app.
const Name = "redis"

// KeyPass is the settings key to the redis password.
const KeyPass = "redis.pass"

// Redis is the redis app.
type Redis struct {
	core homeapp.Core
}

// New creates a new redis app.
func New(c homeapp.Core) *Redis { return &Redis{core: c} }

func (r *Redis) writeConfig(cont *dock.Cont, pwd string) error {
	confFile := tarutil.NewStream()
	confContent := fmt.Sprintf("requirepass %q\n", pwd)
	confFile.AddString("redis.conf", &tarutil.Meta{
		Mode:    0700,
		UserID:  999, // redis docker's uid and gid.
		GroupID: 999,
	}, confContent)
	return dock.CopyInTarStream(cont, confFile, "/etc")
}

func (r *Redis) cont() *dock.Cont {
	return dock.NewCont(r.core.Docker(), homeapp.Cont(r.core, Name))
}

func (r *Redis) createCont(image, pwd string) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}
	if pwd == "" {
		return nil, errcode.InvalidArgf("redis password empty")
	}

	config := &dock.ContConfig{
		Name:          homeapp.Cont(r.core, Name),
		Network:       homeapp.Network(r.core),
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Cmd:           []string{"redis-server", "/etc/redis.conf"},
		Labels:        drvcfg.NewNameLabel(Name),
	}

	cont, err := dock.CreateCont(r.core.Docker(), image, config)
	if err != nil {
		return nil, errcode.Annotate(err, "create docker")
	}
	if err := r.writeConfig(cont, pwd); err != nil {
		cont.Drop()
		return nil, errcode.Annotate(err, "write config file")
	}

	return cont, nil
}

func (r *Redis) install(image string) error {
	pwd, err := r.password()
	if err != nil {
		return errcode.Annotate(err, "read password")
	}
	cont, err := r.createCont(image, pwd)
	if err != nil {
		return err
	}
	if err := cont.Start(); err != nil {
		return errcode.Annotate(err, "start container")
	}
	return nil
}

func (r *Redis) password() (string, error) {
	return apputil.ReadPasswordOrSetRandom(r.core.Settings(), KeyPass)
}

// Change changes the app's version.
func (r *Redis) Change(from, to *drvapi.AppMeta) error {
	if from != nil {
		if err := apputil.DropIfExists(r.cont()); err != nil {
			return errcode.Annotate(err, "drop old redis container")
		}
	}

	if to == nil {
		vol := homeapp.Vol(r.core, Name)
		if err := dock.RemoveVolume(r.core.Docker(), vol); err != nil {
			return errcode.Annotate(err, "remove volume")
		}
		return nil
	}

	return r.install(homeapp.Image(to))
}

// Start starts the app.
func (r *Redis) Start() error { return r.cont().Start() }

// Stop stops the app.
func (r *Redis) Stop() error { return r.cont().Stop() }
