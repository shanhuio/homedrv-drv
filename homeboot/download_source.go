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
	"io"
	"path"

	"shanhu.io/homedrv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
)

// FetchChannelRelease fetch the release from a particular channel.
func FetchChannelRelease(c *httputil.Client, ch string) (
	*drvapi.Release, error,
) {
	r := new(drvapi.Release)
	const p = "/pubapi/release/channel"
	if err := c.Call(p, ch, r); err != nil {
		return nil, errcode.Annotate(err, "fetch channel")
	}
	return r, nil
}

// FetchBuildRelease fetches the a particular build.
func FetchBuildRelease(c *httputil.Client, b string) (*drvapi.Release, error) {
	r := new(drvapi.Release)
	const p = "/pubapi/release/get"
	if err := c.Call(p, b, r); err != nil {
		return nil, errcode.Annotate(err, "fetch release")
	}
	return r, nil
}

// DownloadSource is a source for downloading a release.
type DownloadSource struct {
	Build      func(b string) (*drvapi.Release, error)
	Channel    func(ch string) (*drvapi.Release, error)
	OpenObject func(name string) (io.ReadCloser, error)
	OpenDocker func(name, hash string) (io.ReadCloser, error)
}

// OfficialDownloadSource creates a downloader downloading from
// HomeDrive official website.
func OfficialDownloadSource(c *httputil.Client) *DownloadSource {
	return &DownloadSource{
		Build: func(b string) (*drvapi.Release, error) {
			return FetchBuildRelease(c, b)
		},
		Channel: func(ch string) (*drvapi.Release, error) {
			return FetchChannelRelease(c, ch)
		},
		OpenObject: func(name string) (io.ReadCloser, error) {
			p := path.Join("/dl/obj", name)
			req, err := c.Get(p)
			if err != nil {
				return nil, err
			}
			return req.Body, nil
		},
		OpenDocker: func(name, hash string) (io.ReadCloser, error) {
			p := path.Join("/dl/docker", name, hash+".tar.gz")
			req, err := c.Get(p)
			if err != nil {
				return nil, err
			}
			return req.Body, nil
		},
	}
}
