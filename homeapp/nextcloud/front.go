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

package nextcloud

import (
	"shanhu.io/homedrv/drv/drvapi"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
	"shanhu.io/homedrv/drv/homeapp"
	"shanhu.io/homedrv/drv/homeapp/apputil"
	"shanhu.io/pub/dock"
	"shanhu.io/pub/errcode"
)

// Front is the ncfront app.
type Front struct {
	core homeapp.Core
}

// NewFront creates a new ncfront app.
func NewFront(c homeapp.Core) *Front { return &Front{core: c} }

func (f *Front) cont() *dock.Cont {
	return dock.NewCont(f.core.Docker(), homeapp.Cont(f.core, NameFront))
}

func (f *Front) createCont(image string) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}

	nextcloudAddr := homeapp.Cont(f.core, Name) + ":80"
	config := &dock.ContConfig{
		Name:          homeapp.Cont(f.core, NameFront),
		Network:       homeapp.Network(f.core),
		Env:           map[string]string{"NEXTCLOUD": nextcloudAddr},
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Labels:        drvcfg.NewNameLabel(NameFront),
	}
	return dock.CreateCont(f.core.Docker(), image, config)
}

func (f *Front) startWithImage(image string) error {
	cont, err := f.createCont(image)
	if err != nil {
		return errcode.Annotate(err, "create ncfront container")
	}
	return cont.Start()
}

func (f *Front) install(image string) error {
	return f.startWithImage(image)
}

// Start starts the app.
func (f *Front) Start() error { return f.cont().Start() }

// Stop stops the app.
func (f *Front) Stop() error { return f.cont().Stop() }

// Change changes the app's version.
func (f *Front) Change(from, to *drvapi.AppMeta) error {
	if from != nil {
		if err := apputil.DropIfExists(f.cont()); err != nil {
			return errcode.Annotate(err, "drop old ncfront container")
		}
	}
	if to == nil {
		return nil
	}
	return f.install(homeapp.Image(to))
}
