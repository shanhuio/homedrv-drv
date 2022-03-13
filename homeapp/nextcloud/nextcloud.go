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

package nextcloud

import (
	"io"
	"log"
	"time"

	"shanhu.io/homedrv/drvapi"
	"shanhu.io/homedrv/homeapp"
	"shanhu.io/homedrv/homeapp/apputil"
	"shanhu.io/homedrv/homeapp/postgres"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/semver"
	"shanhu.io/virgo/dock"
)

// Nextcloud is the Nextcloud app.
type Nextcloud struct {
	core homeapp.Core
}

// New creates a new Nextcloud app.
func New(c homeapp.Core) *Nextcloud { return &Nextcloud{core: c} }

func (n *Nextcloud) cont() *dock.Cont {
	cont := homeapp.Cont(n.core, Name)
	return dock.NewCont(n.core.Docker(), cont)
}

func (n *Nextcloud) startWithImage(image string, config *config) error {
	return start(n.core, image, config)
}

func (n *Nextcloud) fix() error {
	version, err := readTrueVersion(n.cont())
	if err != nil {
		return errcode.Annotate(err, "get version")
	}
	major, err := semver.Major(version)
	if err != nil {
		return errcode.Add(errcode.Internal, err)
	}
	return n.fixVersion(major)
}

func (n *Nextcloud) fixVersion(major int) error {
	cont := n.cont()

	if major >= 21 {
		// For version 21+, this needs to be executed every time a new
		// docker is installed.
		if err := aptUpdate(cont, io.Discard); err != nil {
			return errcode.Annotate(err, "apt update for nc21")
		}

		pkgs := []string{
			"libmagickcore-6.q16-6-extra",
			"smbclient",
			"libsmbclient-dev",
		}
		if err := aptInstall(cont, pkgs, io.Discard); err != nil {
			return errcode.Annotate(err, "install additional packages")
		}

		if err := enableSMB(cont, io.Discard); err != nil {
			return errcode.Annotate(err, "enable SMB")
		}
	}

	if err := setCronMode(cont); err != nil {
		return errcode.Annotate(err, "set cron mode")
	}

	// Might be also needed in minor upgrades.
	const addIndexCmd = "db:add-missing-indices"
	if _, err := occOutput(cont, []string{addIndexCmd, "-n"}); err != nil {
		return errcode.Annotate(err, addIndexCmd)
	}

	k := fixKey(major)
	if k == "" {
		return nil
	}
	settings := n.core.Settings()
	ok, err := settings.Has(k)
	if err != nil {
		return errcode.Annotatef(err, "check fixed flag v%d", major)
	}
	if ok {
		return nil
	}

	for _, cmd := range []string{
		"db:convert-filecache-bigint",
		"db:add-missing-columns",
		"db:add-missing-primary-keys",
	} {
		if _, err := occOutput(cont, []string{cmd, "-n"}); err != nil {
			return errcode.Annotate(err, cmd)
		}
	}

	if err := settings.Set(k, true); err != nil {
		return errcode.Annotatef(err, "set fixed flag v%d", major)
	}
	return nil
}

func (n *Nextcloud) upgrade(
	img string, from *drvapi.AppMeta,
	ladder []*drvapi.StepVersion, config *config,
) error {
	if len(ladder) == 0 {
		return errcode.InvalidArgf("nextcloud ladder missing")
	}

	var defVersion string
	if from != nil {
		defVersion = from.SemVersion
	}
	version, err := readVersion(n.cont(), defVersion)
	if err != nil {
		return errcode.Annotatef(err, "read version")
	}
	if version == "" {
		return errcode.Annotatef(err, "cannot determin last version")
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

func (n *Nextcloud) upgrade1(img, ver string, c *config) error {
	cont := n.cont()
	if err := apputil.DropIfExists(cont); err != nil {
		return errcode.Annotate(err, "drop container")
	}
	// This is a dangerous moment. If the machine restarts at this point,
	// nextcloud won't be there anymore.
	if err := n.startWithImage(img, c); err != nil {
		return errcode.Annotate(err, "start new nextcloud")
	}
	if err := waitReady(cont, 5*time.Minute, ver); err != nil {
		return errcode.Annotate(err, "wait for install complete")
	}
	if err := setRedisPassword(cont, c.redisPassword); err != nil {
		return errcode.Annotate(err, "set redis password")
	}
	if err := n.fix(); err != nil {
		return errcode.Annotate(err, "fix post-install issues")
	}
	return nil
}

// Start starts the app.
func (n *Nextcloud) Start() error { return n.cont().Start() }

// Stop stops the app.
func (n *Nextcloud) Stop() error { return n.cont().Stop() }

func (n *Nextcloud) config() (*config, error) { return loadConfig(n.core) }

// Change changes the version of the app.
func (n *Nextcloud) Change(from, to *drvapi.AppMeta) error {
	cont := n.cont()
	if to == nil {
		if err := n.registerDomains(nil); err != nil {
			return errcode.Annotate(err, "unregister domains")
		}
		if err := apputil.DropIfExists(cont); err != nil {
			return errcode.Annotate(err, "drop old nextcloud container")
		}
		psql, err := n.db()
		if err != nil {
			return errcode.Annotate(err, "get db handle")
		}
		if err := psql.DropDB(Name); err != nil {
			return errcode.Annotate(err, "drop nextcloud db")
		}
		vol := homeapp.Vol(n.core, Name)
		if err := dock.RemoveVolume(n.core.Docker(), vol); err != nil {
			return errcode.Annotate(err, "remove volume")
		}
		return nil
	}

	config, err := n.config()
	if err != nil {
		return errcode.Annotate(err, "load config")
	}
	if from == nil {
		return n.install(homeapp.Image(to), config)
	}
	return n.upgrade(homeapp.Image(to), from, to.Steps, config)
}

func (n *Nextcloud) db() (*postgres.Postgres, error) {
	db, err := n.core.App(postgres.Name)
	if err != nil {
		return nil, errcode.Annotate(err, "reflect postgres db")
	}
	psql, ok := db.(*postgres.Postgres)
	if !ok {
		return nil, errcode.Internalf("reflected db is not postgres")
	}
	return psql, nil
}

func (n *Nextcloud) registerDomains(domains []string) error {
	appDomains := n.core.Domains()

	if len(domains) == 0 {
		return appDomains.Clear(Name)
	}
	m := &homeapp.DomainMap{
		App: Name,
		Map: make(map[string]*homeapp.DomainEntry),
	}
	ncFrontAddr := homeapp.Cont(n.core, NameFront) + ":8080"
	for _, d := range domains {
		m.Map[d] = &homeapp.DomainEntry{Dest: ncFrontAddr}
	}
	return appDomains.Set(m)
}

func (n *Nextcloud) install(image string, config *config) error {
	psql, err := n.db()
	if err != nil {
		return errcode.Annotate(err, "get db handle")
	}
	if err := psql.CreateDB(Name, config.dbPassword); err != nil {
		return errcode.Annotate(err, "create db")
	}

	if err := n.startWithImage(image, config); err != nil {
		return errcode.Annotate(err, "start container")
	}
	cont := n.cont()
	if err := waitReady(cont, 30*time.Minute, ""); err != nil {
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

// Cron runs the nextcloud cron job
func (n *Nextcloud) Cron() error { return cron(n.cont()) }
