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
	"errors"
	"fmt"
	"log"
	"time"

	"shanhu.io/homedrv/drvapi"
	"shanhu.io/homedrv/homeapp/apputil"
	"shanhu.io/homedrv/homeboot"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
)

func updateAppsAndDoorway(d *drive, r *drvapi.Release) error {
	dl, err := downloader(d)
	if err != nil {
		return errcode.Annotate(err, "init downloader")
	}

	// Need to refetch the release info because the one fetched from the
	// last version is unmarshalled by an older core, and the JSON blob might
	// be incomplete.
	refetched, err := dl.FetchBuild(r.Name)
	if err != nil {
		return errcode.Annotate(err, "refetch release info")
	}
	r = refetched

	// And also need to download again.
	dlConfig := &homeboot.DownloadConfig{
		Release:            r,
		Naming:             d.config.Naming,
		CurrentSemVersions: d.apps.semVersions(),
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

type taskUpdate struct {
	drive *drive
	rel   *drvapi.Release
}

func (t *taskUpdate) run() error {
	d := t.drive
	rel := t.rel

	dl, err := downloader(d)
	if err != nil {
		return errcode.Annotate(err, "init downloader")
	}
	config := &homeboot.DownloadConfig{
		Release:            rel,
		Naming:             d.config.Naming,
		CurrentSemVersions: d.apps.semVersions(),
	}
	if _, err := dl.DownloadRelease(config); err != nil {
		return errcode.Annotate(err, "download release")
	}

	if err := d.settings.Set(keyBuildUpdating, rel); err != nil {
		return errcode.Annotatef(err, "set %q", keyBuildUpdating)
	}

	// If update succeeds, the core will be swapped with a new
	// instance, and updateCore() will never return.
	if err := updateCore(d, rel.Jarvis); err != nil {
		if err != apputil.ErrSameImage {
			return errcode.Annotate(err, "update core")
		}
		// Core did not update, finish the rest of the system.
		return updateAppsAndDoorway(t.drive, t.rel)
	}

	// This point should be unreachable.
	return errcode.Internalf("core updated but still returned")
}

func pushManualUpdate(d *drive, relBytes []byte) error {
	if err := d.settings.Set(keyManualBuild, relBytes); err != nil {
		return errcode.Annotate(err, "set to local build mode")
	}
	go func() {
		if err := updateDriveToManualBuild(d); err != nil {
			log.Printf("push update failed: %q", err)
		}
	}()
	return nil
}

func updateDriveToManualBuild(d *drive) error {
	r, err := readManualBuild(d)
	if err != nil {
		return errcode.Annotate(err, "read manual build release")
	}
	t := &taskUpdate{drive: d, rel: r}
	return d.tasks.run("update to custom build", t)
}

var errAlreadyUpToDate = errors.New("already up to date")

func queryUpdate(c *httputil.Client, ch, cur, tags string, manual bool) (
	*drvapi.Release, error,
) {
	req := &drvapi.UpdateQueryRequest{
		Channel:      ch,
		CurrentBuild: cur,
		Tags:         tags,
		Manual:       manual,
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

func updateDriveOnChannel(d *drive, ch string, manual bool) error {
	tags := d.tags()

	cur := new(drvapi.Release)
	if err := d.settings.Get(keyBuild, cur); err != nil {
		return errcode.Annotate(err, "fetch current build")
	}

	client, err := d.dialServer()
	if err != nil {
		return errcode.Annotate(err, "dial server")
	}
	r, err := queryUpdate(client, ch, cur.Name, tags, manual)
	if err != nil {
		if err == errAlreadyUpToDate {
			return err
		}
		return errcode.Annotate(err, "query channel update")
	}
	t := &taskUpdate{drive: d, rel: r}
	return d.tasks.run(fmt.Sprintf("update to release %q", r.Name), t)
}

func sleepOrSignal(d time.Duration, signal <-chan bool) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-signal:
		return false
	}
}

func cronUpdateOnChannel(d *drive, signal <-chan bool) {
	defer log.Println("cron update on channel exited")

	// TODO(h8liu): add a stop signal, so this background update
	// routine can be stopped gracefully.

	ch := d.downloadConfig().Channel
	if ch == "" {
		log.Println("not running on a channel, quiting update cron")
		return
	}

	const interval = time.Minute * 10
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tickerChan := ticker.C

	manual := false
	stop := false
	for !stop {
		if mb, err := d.settings.Has(keyManualBuild); err != nil {
			log.Println(errcode.Annotate(err, "check manual build"))
		} else if mb {
			log.Println("on manual build mode; no cron update needed")
			return
		}

		for {
			const errInterval = time.Minute * 5
			if err := updateDriveOnChannel(d, ch, manual); err != nil {
				if err == errAlreadyUpToDate {
					break
				}
				log.Printf("update homedrive: %s", err)
				sleepOrSignal(errInterval, signal)
				continue
			}
			break
		}

		manual = false
		select {
		case <-tickerChan:
		case b := <-signal:
			if !b {
				stop = true
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
	return updateAppsAndDoorway(d, r)
}
