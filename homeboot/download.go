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
	"log"
	"path"
	"strconv"

	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/virgo/dock"
)

// Downloader is a downloader for downloading docker images.
type Downloader struct {
	client *httputil.Client
	dock   *dock.Client
}

// NewDownloader creates a new downloader.
func NewDownloader(client *httputil.Client, dock *dock.Client) *Downloader {
	return &Downloader{client: client, dock: dock}
}

// DownloadImage downloads a docker image.
func (d *Downloader) DownloadImage(p string) error {
	req, err := d.client.Get(p)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	return dock.LoadImages(d.dock, req.Body)
}

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

// FetchRelease fethes the release specified by the InstallConfig
func FetchRelease(c *httputil.Client, config *InstallConfig) (
	*drvapi.Release, error,
) {
	if config.Release != nil {
		return config.Release, nil
	}
	if config.Channel != "" {
		return FetchChannelRelease(c, config.Channel)
	}
	if config.Build != "" {
		return FetchBuildRelease(c, config.Build)
	}
	return nil, errcode.InvalidArgf("no build specified")
}

// DownloadRelease downloads an entire release.
func (d *Downloader) DownloadRelease(c *InstallConfig) (
	*drvapi.Release, error,
) {
	r, err := FetchRelease(d.client, c)
	if err != nil {
		return nil, errcode.Annotate(err, "fetch release")
	}
	type image struct{ name, repo, tag, hash string }
	images := []*image{{
		name: "jarvis",
		hash: r.Jarvis,
		repo: drvcfg.Image(c.Naming, "core"),
	}}
	if !c.CoreOnly {
		images = append(images, &image{name: "doorway", hash: r.Doorway})

		images = append(images, []*image{
			{name: "redis", hash: r.Redis},
			{name: "postgres", hash: r.Postgres},
		}...)

		images = append(images, &image{name: "ncfront", hash: r.NCFront})
		// TODO(h8liu): do not download the full ladder every time.
		for _, nc := range r.Nextclouds {
			images = append(images, &image{
				name: "nextcloud",
				tag:  strconv.Itoa(nc.Major),
				hash: nc.Image,
			})
		}
	}

	for _, img := range images {
		if img.hash == "" {
			continue // Hash missing, just skip.
		}
		found, err := dock.HasImage(d.dock, img.hash)
		if err != nil {
			return nil, errcode.Annotatef(err, "check image %q", img.name)
		}
		display := img.name
		if img.tag != "" {
			display = fmt.Sprintf("%s:%s", img.name, img.tag)
		}
		if !found {
			if r.ImageSums == nil {
				log.Printf("downloading image %q", display)
				p := path.Join("/dl/docker", img.name, img.hash+".tar.gz")
				if err := d.DownloadImage(p); err != nil {
					return nil, errcode.Annotatef(
						err, "download image %q", display,
					)
				}
			} else {
				obj, ok := r.ImageSums[img.hash]
				if !ok {
					return nil, errcode.InvalidArgf(
						"object for image %q missing", display,
					)
				}
				p := path.Join("/dl/obj", obj)
				if err := d.DownloadImage(p); err != nil {
					return nil, errcode.Annotatef(
						err, "download image %q", display,
					)
				}
			}
		}
		repo := img.repo
		if repo == "" {
			repo = drvcfg.Image(c.Naming, img.name)
		}
		tag := img.tag
		if tag == "" {
			tag = "main"
		}
		log.Printf("tag image as %s:%s", repo, tag)
		if err := dock.TagImage(d.dock, img.hash, repo, tag); err != nil {
			return nil, errcode.Annotatef(
				err, "tag image %q", display,
			)
		}
	}

	return r, nil
}
