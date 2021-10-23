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
	"io"
	"log"
	"strconv"

	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/virgo/dock"
)

// Downloader is a downloader for downloading docker images.
type Downloader struct {
	src  *DownloadSource
	dock *dock.Client
}

// NewOfficialDownloader creates a new downloader that downloads
// from the official source.
func NewOfficialDownloader(
	c *httputil.Client, dock *dock.Client,
) *Downloader {
	src := OfficialDownloadSource(c)
	return NewDownloader(src, dock)
}

// NewDownloader creates a new downloader.
func NewDownloader(src *DownloadSource, dock *dock.Client) *Downloader {
	return &Downloader{src: src, dock: dock}
}

func (d *Downloader) loadImage(r io.ReadCloser, err error) error {
	if err != nil {
		return err
	}
	defer r.Close()
	return dock.LoadImages(d.dock, r)
}

// FetchBuild fetches a particular build.
func (d *Downloader) FetchBuild(b string) (*drvapi.Release, error) {
	return d.src.Build(b)
}

func (d *Downloader) fetchRelease(config *DownloadConfig) (
	*drvapi.Release, error,
) {
	if config.Release != nil {
		return config.Release, nil
	}
	if config.Channel != "" {
		return d.src.Channel(config.Channel)
	}
	if config.Build != "" {
		return d.src.Build(config.Build)
	}
	return nil, errcode.InvalidArgf("no build specified")
}

type downloadImage struct {
	name, repo, tag, hash string
}

func (d *Downloader) downloadImage(
	img *downloadImage, display string,
	sums map[string]string,
) error {
	log.Printf("downloading image %q", display)
	if sums == nil {
		return d.loadImage(d.src.OpenDocker(img.name, img.hash))
	}

	obj, ok := sums[img.hash]
	if !ok {
		return errcode.InvalidArgf(
			"object for image %q missing", display,
		)
	}
	return d.loadImage(d.src.OpenObject(obj))
}

func (d *Downloader) downloadImages(
	imgs []*downloadImage, naming *drvcfg.Naming,
	sums map[string]string,
) error {
	for _, img := range imgs {
		if img.hash == "" {
			continue // Hash missing, just skip.
		}
		display := img.name
		if img.tag != "" {
			display = fmt.Sprintf("%s:%s", img.name, img.tag)
		}

		found, err := dock.HasImage(d.dock, img.hash)
		if err != nil {
			return errcode.Annotatef(err, "check image %q", img.name)
		}
		if !found {
			if err := d.downloadImage(img, display, sums); err != nil {
				return errcode.Annotatef(err, "download %q", display)
			}
		}

		repo := drvcfg.Image(naming, img.name)
		tag := img.tag
		if tag == "" {
			tag = "main"
		}
		log.Printf("tag image as %s:%s", repo, tag)
		if err := dock.TagImage(d.dock, img.hash, repo, tag); err != nil {
			return errcode.Annotatef(err, "tag image %q", display)
		}
	}
	return nil
}

// DownloadRelease downloads an entire release.
func (d *Downloader) DownloadRelease(c *DownloadConfig) (
	*drvapi.Release, error,
) {
	r, err := d.fetchRelease(c)
	if err != nil {
		return nil, errcode.Annotate(err, "fetch release")
	}
	images := []*downloadImage{{
		name: "jarvis",
		hash: r.Jarvis,
	}}
	if !c.CoreOnly {
		images = append(images, &downloadImage{
			name: "doorway", hash: r.Doorway,
		})

		images = append(images, []*downloadImage{
			{name: "redis", hash: r.Redis},
			{name: "postgres", hash: r.Postgres},
		}...)

		images = append(images, &downloadImage{
			name: "ncfront", hash: r.NCFront,
		})

		for i, nc := range r.Nextclouds {
			if c.LatestOnly && i != len(r.Nextclouds)-1 {
				continue
			}
			images = append(images, &downloadImage{
				name: "nextcloud",
				tag:  strconv.Itoa(nc.Major),
				hash: nc.Image,
			})
		}
	}

	if err := d.downloadImages(images, c.Naming, r.ImageSums); err != nil {
		return nil, err
	}

	return r, nil
}
