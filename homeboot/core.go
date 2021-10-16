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
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/osutil"
	"shanhu.io/misc/tarutil"
	"shanhu.io/virgo/dock"
)

// CoreMount is the mount point of jarvis volume.
const CoreMount = "/opt/jarvis/var"

// CoreConfig specifies how to start a core.
type CoreConfig struct {
	Drive *drvcfg.Config
	Image string
	Files *tarutil.Stream

	BindSysDock bool // bind system-docker.sock
}

// StartCore starts the core.homedrv container.
func StartCore(client *dock.Client, config *CoreConfig) (string, error) {
	naming := config.Drive.Naming
	image := config.Image
	if image == "" {
		image = drvcfg.Image(naming, "core")
	}
	name := drvcfg.Core(naming)
	labels := drvcfg.NewNameLabel("core")

	binds := []*dock.ContMount{{
		Type: dock.MountVolume,
		Host: name,
		Cont: CoreMount,
	}}

	bindSocks := []string{dock.Socket}
	if config.BindSysDock {
		bindSocks = append(bindSocks, systemDockSock)
	}
	for _, s := range bindSocks {
		ok, err := osutil.IsSock(s)
		if err != nil {
			return "", errcode.Annotatef(err, "check socket %q", s)
		}
		if !ok {
			return "", errcode.Annotatef(err, "socket %q", s)
		}
		binds = append(binds, &dock.ContMount{Host: s, Cont: s})
	}

	dockConfig := &dock.ContConfig{
		Name:        name,
		Network:     drvcfg.Network(naming),
		AutoRestart: true,
		Mounts:      binds,
		Labels:      labels,
	}
	// Note that core cannot bind ports, either TCP or UDP,
	// because updating the core needs to run a new one along
	// side with the old one.

	if _, err := dock.CreateVolumeIfNotExist(
		client, name, &dock.VolumeConfig{Labels: labels},
	); err != nil {
		return "", errcode.Annotate(err, "create volume for core")
	}

	cont, err := dock.CreateCont(client, image, dockConfig)
	if err != nil {
		return "", errcode.Annotate(err, "create core container")
	}

	if config.Files != nil {
		if err := dock.CopyInTarStream(
			cont, config.Files, CoreMount,
		); err != nil {
			cont.Drop()
			return "", errcode.Annotate(err, "copy in init files")
		}
	}

	if err := cont.Start(); err != nil {
		return "", errcode.Annotate(err, "start core container")
	}
	return cont.ID(), nil
}
