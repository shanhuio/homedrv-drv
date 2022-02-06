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
	"time"

	"shanhu.io/homedrv/homeapp/nextcloud"
	"shanhu.io/misc/errcode"
)

type taskNextcloudCron struct{ drive *drive }

func (t *taskNextcloudCron) run() error {
	d := t.drive
	if !d.apps.isInstalled(nextcloud.Name) {
		return nil
	}
	stub, err := d.apps.stub(nextcloud.Name)
	if err != nil {
		return errcode.Annotate(err, "get nextcloud stub")
	}
	nc, ok := stub.App.(*nextcloud.Nextcloud)
	if !ok {
		return errcode.Internalf("nextcloud stub is %T", stub)
	}
	return nc.Cron()
}

func cronNextcloud(d *drive) {
	const cronPeriod = 5 * time.Minute
	ticker := time.NewTicker(cronPeriod)
	defer ticker.Stop()
	for range ticker.C {
		t := &taskNextcloudCron{drive: d}
		if err := d.tasks.run("nextcloud cron", t); err != nil {
			log.Printf("nextcloud cron: %s", err)
		}
	}
}
