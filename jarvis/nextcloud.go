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
	"io"
	"log"
	"time"

	"shanhu.io/homedrv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/semver"
	"shanhu.io/virgo/dock"
)

type nextcloud struct {
	*drive
}

func newNextcloud(d *drive) *nextcloud { return &nextcloud{drive: d} }

func (n *nextcloud) cont() *dock.Cont {
	return dock.NewCont(n.dock, n.drive.cont(nameNextcloud))
}

func (n *nextcloud) occRet(args []string, out io.Writer) (int, error) {
	return nextcloudOCCRet(n.cont(), args, out)
}

func (n *nextcloud) occ(args []string, out io.Writer) error {
	return nextcloudOCC(n.cont(), args, out)
}

func (n *nextcloud) occOutput(args []string) ([]byte, error) {
	return nextcloudOCCOutput(n.cont(), args)
}

func (n *nextcloud) status() (*nextcloudStatus, int, error) {
	return nextcloudReadStatus(n.cont())
}

func (n *nextcloud) startWithImage(
	image string, config *nextcloudConfig,
) error {
	return nextcloudStart(n.drive, n.dock, image, config)
}

func (n *nextcloud) waitReady(timeout time.Duration, v string) error {
	return nextcloudWaitReady(n.cont(), timeout, v)
}

func (n *nextcloud) version() (string, error) {
	status, ret, err := n.status()
	if err != nil {
		return "", errcode.Annotate(err, "read status")
	}
	if ret != 0 {
		return "", errcode.Internalf("status exit with: %d", ret)
	}
	return status.VersionString, nil
}

func (n *nextcloud) fix() error {
	version, err := n.version()
	if err != nil {
		return errcode.Annotate(err, "get version")
	}
	major, err := semver.Major(version)
	if err != nil {
		return errcode.Add(errcode.Internal, err)
	}
	return n.fixVersion(major)
}

func (n *nextcloud) fixVersion(major int) error {
	// For version 21+, this needs to be executed every time a new
	// docker is installed.
	if major >= 21 {
		cont := n.cont()
		if err := nextcloudAptUpdate(cont, io.Discard); err != nil {
			return errcode.Annotate(err, "apt update for nc21")
		}
		const pkg = "libmagickcore-6.q16-6-extra"
		if err := nextcloudAptInstall(cont, pkg, io.Discard); err != nil {
			return errcode.Annotate(err, "install svg support")
		}
	}

	k := nextcloudFixKey(major)
	if k == "" {
		return nil
	}
	ok, err := n.settings.Has(k)
	if err != nil {
		return errcode.Annotatef(err, "check fixed flag v%d", major)
	}
	if ok {
		return nil
	}

	cont := n.cont()

	for _, cmd := range []string{
		"db:add-missing-indices",
		"db:convert-filecache-bigint",
		"db:add-missing-columns",
		"db:add-missing-primary-keys",
	} {
		if _, err := nextcloudOCCOutput(
			cont, []string{cmd, "-n"},
		); err != nil {
			return errcode.Annotate(err, cmd)
		}
	}

	if err := n.settings.Set(k, true); err != nil {
		return errcode.Annotatef(err, "set fixed flag v%d", major)
	}
	return nil
}

func (n *nextcloud) upgrade(
	img string, ladder []*drvapi.StepVersion, config *nextcloudConfig,
) error {
	if len(ladder) == 0 {
		return errcode.InvalidArgf("nextcloud ladder missing")
	}

	version, err := n.version()
	if err != nil {
		return errcode.Annotate(err, "read version")
	}
	curMajor, err := semver.Major(version)
	if err != nil {
		return errcode.Add(errcode.Internal, err)
	}
	if curMajor < 20 {
		return errcode.Internalf("version <20 not supported: %q", version)
	}

	ladderMap := make(map[int]*drvapi.StepVersion)
	for _, v := range ladder {
		ladderMap[v.Major] = v
	}
	last := ""
	for {
		v, ok := ladderMap[curMajor]
		if !ok { // Top of the upgrade ladder now.
			break
		}
		log.Printf("upgrade nextcloud from %q to %q", version, v.Version)
		if err := n.upgrade1(v.Image, v.Version, config); err != nil {
			return errcode.Annotatef(
				err, "upgrade nextcloud from %q to %q", version, v.Version,
			)
		}
		version = v.Version
		curMajor++
		last = v.Image
		log.Printf("upgraded to %q", version)
	}
	if img != last {
		log.Println("end of upgrade ladder is not target version")
	}
	if err := n.registerDomains(config.domains); err != nil {
		return errcode.Annotate(err, "register nextcloud domains")
	}
	return nil
}

func (n *nextcloud) upgrade1(img, ver string, c *nextcloudConfig) error {
	if err := n.cont().Drop(); err != nil {
		return errcode.Annotatef(err, "drop nextcloud")
	}

	// This is a dangerous moment. If the machine restarts at this point,
	// nextcloud won't be there anymore.
	if err := n.startWithImage(img, c); err != nil {
		return errcode.Annotate(err, "start new nextcloud")
	}
	if err := n.waitReady(5*time.Minute, ver); err != nil {
		return errcode.Annotate(err, "wait for install complete")
	}
	if err := n.setRedisPassword(c.redisPassword); err != nil {
		return errcode.Annotate(err, "set redis password")
	}
	if err := n.fix(); err != nil {
		return errcode.Annotate(err, "fix post-install issues")
	}
	return nil
}

func (n *nextcloud) start() error { return n.cont().Start() }
func (n *nextcloud) stop() error  { return n.cont().Stop() }

func (n *nextcloud) config() (*nextcloudConfig, error) {
	return loadNextcloudConfig(n.drive)
}

func (n *nextcloud) change(from, to *drvapi.AppMeta) error {
	if to == nil {
		if err := n.registerDomains(nil); err != nil {
			return errcode.Annotate(err, "unregister domains")
		}
		if err := n.cont().Drop(); err != nil {
			return errcode.Annotate(err, "drop old nextcloud container")
		}
		psql, err := n.db()
		if err != nil {
			return errcode.Annotate(err, "get db handle")
		}
		if err := psql.dropDB(nameNextcloud); err != nil {
			return errcode.Annotate(err, "drop nextcloud db")
		}
		vol := n.drive.vol(nameNextcloud)
		if err := dock.RemoveVolume(n.dock, vol); err != nil {
			return errcode.Annotate(err, "remove volume")
		}
		return nil
	}

	config, err := n.config()
	if err != nil {
		return errcode.Annotate(err, "load config")
	}
	if from == nil {
		return n.install(appImage(to), config)
	}
	return n.upgrade(appImage(to), to.Steps, config)
}

func (n *nextcloud) db() (*postgres, error) {
	db, err := n.appReflect(namePostgres)
	if err != nil {
		return nil, errcode.Annotate(err, "reflect postgres db")
	}
	psql, ok := db.(*postgres)
	if !ok {
		return nil, errcode.Internalf("reflected db is not postgres")
	}
	return psql, nil
}

func (n *nextcloud) registerDomains(domains []string) error {
	if len(domains) == 0 {
		return n.appDomains.clear(nameNextcloud)
	}
	m := &appDomainMap{
		App: nameNextcloud,
		Map: make(map[string]*appDomainEntry),
	}
	ncFrontAddr := n.drive.cont(nameNCFront) + ":8080"
	for _, d := range domains {
		m.Map[d] = &appDomainEntry{Dest: ncFrontAddr}
	}
	return n.appDomains.set(m)
}

func (n *nextcloud) install(image string, config *nextcloudConfig) error {
	psql, err := n.db()
	if err != nil {
		return errcode.Annotate(err, "get db handle")
	}
	if err := psql.createDB(nameNextcloud, config.dbPassword); err != nil {
		return errcode.Annotate(err, "create db")
	}

	if err := n.startWithImage(image, config); err != nil {
		return errcode.Annotate(err, "start container")
	}
	if err := n.waitReady(30*time.Minute, ""); err != nil {
		return err
	}
	if err := n.fix(); err != nil {
		return errcode.Annotate(err, "fix issues")
	}
	if err := n.registerDomains(config.domains); err != nil {
		return errcode.Annotate(err, "register domains")
	}
	return nil
}

func (n *nextcloud) setRedisPassword(pwd string) error {
	// TODO(h8liu): should first check if redis password is incorrect.
	args := []string{
		"config:system:set", "--quiet",
		"--value=" + pwd,    // value
		"redis", "password", // key
	}
	return n.occ(args, nil)
}
