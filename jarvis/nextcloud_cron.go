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
