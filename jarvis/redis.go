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
	"fmt"
	"log"

	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/homedrv/homeapp"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/tarutil"
	"shanhu.io/virgo/dock"
)

type redis struct {
	core homeapp.Core
}

func newRedis(c homeapp.Core) *redis { return &redis{core: c} }

func (r *redis) writeConfig(cont *dock.Cont, pwd string) error {
	confFile := tarutil.NewStream()
	confContent := fmt.Sprintf("requirepass %q\n", pwd)
	confFile.AddString("redis.conf", &tarutil.Meta{
		Mode:    0700,
		UserID:  999, // redis docker's uid and gid.
		GroupID: 999,
	}, confContent)
	return dock.CopyInTarStream(cont, confFile, "/etc")
}

func (r *redis) cont() *dock.Cont {
	return dock.NewCont(r.core.Docker(), appCont(r.core, nameRedis))
}

func (r *redis) createCont(image, pwd string) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}
	if pwd == "" {
		return nil, errcode.InvalidArgf("redis password empty")
	}

	config := &dock.ContConfig{
		Name:          appCont(r.core, nameRedis),
		Network:       appNetwork(r.core),
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Cmd:           []string{"redis-server", "/etc/redis.conf"},
		Labels:        drvcfg.NewNameLabel(nameRedis),
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

func (r *redis) install(image string) error {
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

func (r *redis) update(image string, force bool) error {
	if image == "" {
		return errcode.InvalidArgf("redis image empty")
	}
	contName := appCont(r.core, nameRedis)
	d := r.core.Docker()
	if !force {
		if err := dropContIfDifferent(d, contName, image); err != nil {
			if err == errSameImage {
				return nil
			}
			return err
		}
	} else {
		c := dock.NewCont(d, contName)
		if err := c.Drop(); err != nil {
			return errcode.Annotatef(err, "drop redis container")
		}
	}

	log.Println("update redis")
	return r.install(image)
}

func (r *redis) password() (string, error) {
	return readPasswordOrSetRandom(r.core.Settings(), keyRedisPass)
}

func (r *redis) Change(from, to *drvapi.AppMeta) error {
	if from != nil {
		if err := r.cont().Drop(); err != nil {
			return errcode.Annotate(err, "drop old redis container")
		}
	}

	if to == nil {
		vol := appVol(r.core, nameRedis)
		if err := dock.RemoveVolume(r.core.Docker(), vol); err != nil {
			return errcode.Annotate(err, "remove volume")
		}
		return nil
	}

	return r.install(appImage(to))
}

func (r *redis) Start() error { return r.cont().Start() }
func (r *redis) Stop() error  { return r.cont().Stop() }
