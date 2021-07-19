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

package jarvis

import (
	"errors"
	"log"
	"time"

	"shanhu.io/homedrv/drvapi"
	"shanhu.io/homedrv/homeboot"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/virgo/dock"
)

var errAlreadyUpToDate = errors.New("already up to date")

func queryUpdate(c *httputil.Client, ch, build, tags string, manual bool) (
	*drvapi.Release, error,
) {
	req := &drvapi.UpdateQueryRequest{
		Channel:      ch,
		CurrentBuild: build,
		Manual:       manual,
		Tags:         tags,
	}
	resp := new(drvapi.UpdateQueryResponse)
	if err := c.Call("/pubapi/update/query", req, resp); err != nil {
		return nil, err
	}
	if resp.AlreadyLatest {
		return nil, errAlreadyUpToDate
	}
	return resp.Release, nil
}

func updateHomeDrive(
	d *drive, c *homeboot.InstallConfig, manual bool,
) error {
	tags := d.tags()

	d.updateMutex.Lock()
	defer d.updateMutex.Unlock()

	cur := new(drvapi.Release)
	if err := d.settings.Get(keyBuild, cur); err != nil {
		return errcode.Annotate(err, "fetch current build")
	}

	client, err := d.dialServer()
	if err != nil {
		return errcode.Annotate(err, "dial server")
	}
	var toInstall *drvapi.Release
	if c.Build != "" {
		if cur.Name == c.Build {
			return errAlreadyUpToDate
		}
		r, err := homeboot.FetchBuildRelease(client, c.Build)
		if err != nil {
			return errcode.Annotate(err, "fetch build release")
		}
		toInstall = r
	} else if c.Channel != "" {
		r, err := queryUpdate(client, c.Channel, cur.Name, tags, manual)
		if err != nil {
			if err == errAlreadyUpToDate {
				return err
			}
			return errcode.Annotate(err, "query channel update")
		}
		toInstall = r
	} else {
		return errcode.Internalf("not sure how to update")
	}

	dl := homeboot.NewDownloader(client, d.dock)
	newConfig := &homeboot.InstallConfig{
		Release: toInstall,
		Naming:  d.config.Naming,
	}
	if _, err := dl.DownloadRelease(newConfig); err != nil {
		return errcode.Annotate(err, "download release")
	}

	if err := d.settings.Set(keyBuildUpdating, toInstall); err != nil {
		return errcode.Annotatef(err, "set %q", keyBuildUpdating)
	}

	// If update succeeds, the core will be swapped with a new
	// instance, and updateCore() will never return.
	if err := updateCore(d, toInstall.Jarvis); err != nil {
		if err != errSameImage {
			return errcode.Annotate(err, "update core")
		}
		// Core did not update, finish the rest of the system.
		return finishUpdate(d, toInstall)
	}

	// This point should be unreachable.
	return errcode.Internalf("core updated but still returned")
}

func bgUpdate(d *drive, c *homeboot.InstallConfig, requests <-chan string) {
	var tickerChan <-chan time.Time
	if c.Channel != "" {
		const interval = time.Minute * 10
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		tickerChan = ticker.C
	}

	manual := false
	for {
		for {
			const errInterval = time.Minute * 5
			if err := updateHomeDrive(d, c, manual); err != nil {
				if err == errAlreadyUpToDate {
					break
				}
				log.Printf("update homedrive: %s", err)
				time.Sleep(errInterval) // TODO(h8liu): exp-backoff
				continue
			}
			break
		}

		manual = false
		select {
		case <-tickerChan:
		case build := <-requests:
			if c.Channel == "" && build == "" {
				log.Println("cannot update with no build specified")
			} else {
				c.Build = build
			}
			manual = true
		}
	}
}

func maybeFinishUpdate(d *drive) error {
	r := new(drvapi.Release)
	if err := d.settings.Get(keyBuildUpdating, r); err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}
		return errcode.Annotate(err, "get pending installation")
	}
	if r.Name == "" {
		return nil
	}
	return finishUpdate(d, r)
}

func finishUpdate(d *drive, r *drvapi.Release) error {
	client, err := d.dialServer()
	if err != nil {
		return errcode.Annotate(err, "dial server")
	}

	// Need to refetch the release info because the one fetched from the
	// last version is unmarshalled by an older core, and the JSON blob might
	// be incomplete.
	refetched, err := homeboot.FetchBuildRelease(client, r.Name)
	if err != nil {
		return errcode.Annotate(err, "refetch release info")
	}
	r = refetched

	// And also need to download again.
	dl := homeboot.NewDownloader(client, d.dock)
	dlConfig := &homeboot.InstallConfig{
		Release: r,
		Naming:  d.config.Naming,
	}
	if _, err := dl.DownloadRelease(dlConfig); err != nil {
		return errcode.Annotate(err, "download release")
	}

	d.appRegistry.setRelease(r)
	if err := d.apps.update(); err != nil {
		return errcode.Annotate(err, "update apps")
	}

	log.Println("upgrade doorway")
	if err := updateDoorway(d, r.Doorway); err != nil {
		return errcode.Annotate(err, "update doorway")
	}
	if err := updateCleanUp(d, r); err != nil {
		return errcode.Annotate(err, "cleanup")
	}
	var empty drvapi.Release
	if err := d.settings.Set(keyBuild, r); err != nil {
		return errcode.Annotate(err, "set build")
	}
	if err := d.settings.Set(keyBuildUpdating, &empty); err != nil {
		return errcode.Annotate(err, "clear pending")
	}
	log.Println("update complete")
	return nil
}

func updateCleanUp(d *drive, r *drvapi.Release) error {
	// TODO(h8liu): should check image tags, rather than pruning all
	// untagged images.
	opt := &dock.PruneImagesOption{}
	if err := dock.PruneImages(d.dock, opt); err != nil {
		return errcode.Annotate(err, "prune docker images")
	}
	return nil
}

func recreateDoorway(d *drive) error {
	d.updateMutex.Lock()
	defer d.updateMutex.Unlock()

	log.Println("re-creating doorway.")

	c := dock.NewCont(d.dock, d.cont(nameDoorway))
	info, err := c.Inspect()
	if err != nil {
		return errcode.Annotate(err, "inspect current doorway")
	}
	// Force update current doorway to recreate the container.
	if err := updateDoorway(d, info.Image); err != nil {
		return errcode.Annotate(err, "recreate doorway")
	}
	return nil
}
