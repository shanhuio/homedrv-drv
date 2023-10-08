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
	"strings"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/homedrv/drv/drvapi"
)

func releaseImagesToKeep(r *drvapi.Release) map[string]bool {
	m := make(map[string]bool)
	arts := r.Artifacts
	if arts != nil {
		for _, img := range []string{
			arts.Jarvis,
			arts.Doorway,
			arts.Toolbox,
			arts.NCFront,
			arts.Nextcloud,
			arts.Redis,
			arts.Postgres,
		} {
			if img == "" {
				continue
			}
			if !strings.Contains(img, ":") {
				img = "sha256:" + img
			}
			m[img] = true
		}
	}

	for _, app := range r.Apps {
		img := app.Image
		if !strings.Contains(img, ":") {
			img = "sha256:" + img
		}
		m[img] = true
	}
	return m
}

func looksLikeHomeDriveImage(repoTag string) bool {
	for _, prefix := range []string{
		"cr.shanhu.io/",
		"registry.digitalocean.com/shanhu/",
		"cr.homedrive.io/",
		"nextcloud:",
		"postgres:",
		"redis:",
		"ncfront:",
		"core:",
	} {
		if strings.HasPrefix(repoTag, prefix) {
			return true
		}
	}
	return false
}

func updateCleanUp(d *drive, r *drvapi.Release) error {
	keep := releaseImagesToKeep(r)

	images, err := dock.ListImages(d.dock)
	if err != nil {
		return errcode.Annotate(err, "list images")
	}
	removeOpts := &dock.RemoveImageOptions{NoPrune: true}
	for _, img := range images {
		if _, found := keep[img.ID]; found {
			continue
		}
		for _, t := range img.RepoTags {
			if strings.HasPrefix(t, "cr.homedrive.io/empty:") {
				continue
			}
			if strings.HasPrefix(t, "empty:") {
				continue
			}
			if looksLikeHomeDriveImage(t) {
				if err := dock.RemoveImage(
					d.dock, t, removeOpts,
				); err != nil {
					log.Printf("WARNING: untag %q: %s", t, err)
				}
			}
		}
	}

	opt := &dock.PruneImagesOption{}
	if err := dock.PruneImages(d.dock, opt); err != nil {
		return errcode.Annotate(err, "prune untagged docker images")
	}
	return nil
}
