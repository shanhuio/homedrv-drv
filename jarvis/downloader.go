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
	"encoding/json"
	"io"

	"shanhu.io/homedrv/drv/drvapi"
	"shanhu.io/homedrv/drv/homeboot"
	"shanhu.io/pub/errcode"
)

func noOpenDockerInManual(_, _ string) (io.ReadCloser, error) {
	return nil, errcode.InvalidArgf("no docker loading in manual mode")
}

func readManualBuild(d *drive) (*drvapi.Release, error) {
	var bs []byte
	if err := d.settings.Get(keyManualBuild, &bs); err != nil {
		return nil, err
	}
	rel := new(drvapi.Release)
	if err := json.Unmarshal(bs, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

func downloader(d *drive) (*homeboot.Downloader, error) {
	manual, err := d.settings.Has(keyManualBuild)
	if err != nil {
		return nil, errcode.Annotate(err, "check manual build mode")
	}

	getRelease := func(_ string) (*drvapi.Release, error) {
		// When in manual build mode, always returns the release
		// from keyManualBuild.
		return readManualBuild(d)
	}
	if manual {
		src := &homeboot.DownloadSource{
			Build:      getRelease,
			Channel:    getRelease,
			OpenDocker: noOpenDockerInManual,
			OpenObject: func(name string) (io.ReadCloser, error) {
				f, err := d.objects.open(name)
				if err != nil {
					return nil, err
				}
				return f, nil
			},
		}
		return homeboot.NewDownloader(src, d.dock), nil
	}

	if !d.hasServer() {
		return nil, errcode.Internalf("no server to download")
	}
	client, err := d.dialServer()
	if err != nil {
		return nil, errcode.Annotate(err, "dial server")
	}
	return homeboot.NewOfficialDownloader(client, d.dock), nil
}
