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
	"runtime"
	"time"

	"shanhu.io/homedrv/drv/burmilla"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/bosinit"
)

func burmillaSourceInConfig(config *bosinit.Config) string {
	if config == nil {
		return ""
	}
	r := config.Rancher
	if r == nil {
		return ""
	}
	u := r.Upgrade
	if u == nil {
		return ""
	}
	return u.URL
}

func isOSUpdateSupported(d *drive) bool {
	return checkOSUpdateSupported(d) == nil
}

func checkOSUpdateSupported(d *drive) error {
	if !d.hasSys() {
		return errcode.Internalf("this drive does not manage the OS")
	}
	if runtime.GOARCH != "amd64" {
		return errcode.Internalf("os update only supported on amd64")
	}
	return nil
}

func setOSUpdateSource(b *burmilla.Burmilla) error {
	const key = "rancher.upgrade.url"
	current, err := burmilla.ConfigGet(b, key)
	if err != nil {
		return errcode.Annotate(err, "read current config")
	}

	const want = "https://www.homedrive.io/os.yml"
	if current == want {
		return nil
	}

	log.Printf("burmilla os upgrade source changed to: %q", want)
	return burmilla.ConfigSet(b, key, want)
}

func isUEFI(b *burmilla.Burmilla) (bool, error) {
	ret, err := b.ExecRet([]string{"test", "-d", "/sys/firmware/efi"})
	if err != nil {
		return false, errcode.Annotate(err, "execute /bin/test")
	}
	return ret == 0, nil
}

func upgradeBurmillaOS(d *drive, v string) error {
	b, err := d.burmilla()
	if err != nil {
		return errcode.Annotate(err, "init os stub")
	}

	uefi, err := isUEFI(b)
	if err != nil {
		return errcode.Annotate(err, "test if is uefi")
	}

	lines, err := burmilla.ListOS(b)
	if err != nil {
		return errcode.Annotate(err, "list os")
	}

	osList, err := parseOSList(lines)
	if err != nil {
		return errcode.Annotate(err, "parse os list")
	}

	target := osList.find(v)
	if target == nil {
		return errcode.Annotatef(err, "target os version %q not exist", v)
	}
	if target.running {
		return nil // already up-to-date
	}

	running := osList.running()
	if running == nil {
		log.Println("current running os not found")
	} else {
		log.Printf("current running: %s", running.name)
	}

	// Entering danger zone. If power is lost in this process, the box might
	// go into limbo land.

	if _, err := b.ExecOutput([]string{
		"ros", "os", "upgrade", "--no-reboot", "-f", "-i", v,
	}); err != nil {
		return errcode.Annotatef(err, "upgrade burmilla to %q", v)
	}
	if uefi {
		log.Println("updating configs in boot partition")
		if err := runUpdateBootPart(d, v); err != nil {
			return errcode.Annotatef(err, "update boot partition")
		}
		log.Println("boot partition grub config updated")
	}

	log.Println("rebooting machine")
	if _, err := b.ExecOutput([]string{"reboot", "now"}); err != nil {
		return errcode.Annotatef(err, "reboot machine")
	}

	const wait = 3 * time.Minute
	time.Sleep(wait)
	// Should have rebooted now, but just in case it does not
	// return an error here.

	waitSecs := int(wait.Seconds())
	log.Printf("machine not rebooted in %d seconds", waitSecs)
	return errcode.Internalf("machine not rebooted in %d seconds", waitSecs)
}

func waitOSInitDone(b *burmilla.Burmilla) error {
	const n = 5
	for i := 0; i < n; i++ {
		ret1, err1 := b.ExecRet([]string{
			"test", "-f", "/opt/homedrv/.init-done",
		})
		if err1 != nil {
			log.Printf("check .init-done: %s", err1)
		} else if ret1 == 0 {
			return nil
		}

		ret2, err2 := b.ExecRet([]string{
			"test", "-f", "/opt/homedrv/init-done",
		})
		if err2 != nil {
			log.Printf("check init-done: %s", err2)
		} else if ret2 == 0 {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return errcode.Internalf("init-done not found")
}

func switchOSConsole(b *burmilla.Burmilla) error {
	ret, err := b.ExecRet([]string{"which", "apt"})
	if err != nil {
		return errcode.Annotate(err, "find apt")
	}
	if ret == 0 {
		return nil // we get apt, so we are likely already on debian
	}
	if ret != 1 {
		return errcode.Internalf("runs `which apt` returns %d", ret)
	}

	log.Println("switching console to burmilla debian")
	if _, err := b.ExecOutput([]string{
		"ros", "console", "switch", "-f", "default",
	}); err != nil {
		return errcode.Annotate(err, "ros console switch")
	}

	const wait = 3 + time.Minute
	time.Sleep(wait)

	waitSecs := int(wait.Seconds())
	log.Printf("docker not rebooted in %d seconds", waitSecs)
	return errcode.Internalf("docker not rebooted in %d seconds", waitSecs)
}

func updateOS(d *drive) error {
	b, err := d.burmilla()
	if err != nil {
		return errcode.Annotate(err, "init os stub")
	}
	if err := waitOSInitDone(b); err != nil {
		return errcode.Annotate(err, "wait OS init done")
	}
	if err := setOSUpdateSource(b); err != nil {
		return errcode.Annotate(err, "set OS upgrade source")
	}
	const target = "burmilla/os:v1.9.1"
	if err := upgradeBurmillaOS(d, target); err != nil {
		return errcode.Annotate(err, "upgrade OS")
	}
	if err := switchOSConsole(b); err != nil {
		return errcode.Annotate(err, "switch console")
	}
	return nil
}

func maybeUpdateOS(d *drive) error {
	if err := checkOSUpdateSupported(d); err != nil {
		log.Println(err.Error())
		return nil
	}

	updating, err := d.settings.Has(keyBuildUpdating)
	if err != nil {
		return errcode.Annotate(err, "check if it is updating")
	}
	installed, err := d.settings.Has(keyBuild)
	if err != nil {
		return errcode.Annotate(err, "check if installed")
	}
	if installed && !updating {
		// We only upgrade os when it is not installed (first time run) or when
		// it is updating.
		return nil
	}

	return updateOS(d)
}
