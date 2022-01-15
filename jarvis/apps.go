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
	"log"

	"shanhu.io/misc/errcode"
	"shanhu.io/misc/strutil"
)

type apps struct {
	state *appsState
	m     map[string]*appStub

	store   appsStateStore
	querier appQuerier
	maker   appMaker // app stub maker
}

type appsConfig struct {
	store   appsStateStore
	querier appQuerier
	maker   appMaker // optional, can be set later.
}

func newApps(config *appsConfig) (*apps, error) {
	m := make(map[string]*appStub)
	state, err := config.store.load()
	if err != nil {
		return nil, errcode.Annotate(err, "load apps state")
	}

	a := &apps{
		state:   state,
		m:       m,
		store:   config.store,
		querier: config.querier,
	}
	if config.maker != nil {
		if err := a.setMaker(config.maker); err != nil {
			return nil, errcode.Annotate(err, "set maker")
		}
	}
	return a, nil
}

func (a *apps) setMaker(m appMaker) error {
	a.maker = m

	for _, name := range a.state.list() {
		stub, err := m.makeStub(name)
		if err != nil {
			return errcode.Annotatef(err, "make %q stub", name)
		}
		a.m[name] = stub
	}
	return nil
}

func (a *apps) saveState() error { return a.store.save(a.state) }

func (a *apps) stub(name string) (*appStub, error) {
	stub, ok := a.m[name]
	if !ok {
		return nil, errcode.NotFoundf("app %q not found", name)
	}
	return stub, nil
}

func (a *apps) stubOrMake(name string) (*appStub, error) {
	stub, ok := a.m[name]
	if ok {
		return stub, nil
	}
	if a.maker == nil {
		return nil, errcode.Internalf("app maker not set yet")
	}

	stub, err := a.maker.makeStub(name)
	if err != nil {
		return nil, err
	}
	a.m[name] = stub
	return stub, nil
}

func (a *apps) removeStub(name string) { delete(a.m, name) }

func (a *apps) reinstall(name string) error {
	m := a.state.meta(name)
	if m == nil {
		return errcode.NotFoundf("app %q not installed yet", name)
	}

	app, err := a.stub(name)
	if err != nil {
		return errcode.Annotatef(err, "get stub of %q", name)
	}
	return app.change(m, m)
}

func (a *apps) apply(anchored []string) error {
	closure := newAppClosure()
	for _, name := range anchored {
		if _, err := closure.add(a.querier, name, nil); err != nil {
			return errcode.Annotatef(err, "build closure for %q", name)
		}
	}
	keep := closure.appSet()

	uninstall := newAppClosure()
	for _, name := range a.state.list() {
		if _, found := keep[name]; found {
			continue
		}
		if _, err := uninstall.add(a.querier, name, keep); err != nil {
			return errcode.Annotatef(
				err, "build uninstall closure for %q", name,
			)
		}
	}

	for _, floor := range closure.Floors() {
		for _, m := range floor {
			name := m.Name
			old := a.state.meta(name)
			if old != nil {
				if sameAppVersion(old, m) {
					continue // nothing to upgrade
				}
				if old.Version > m.Version {
					return errcode.Internalf(
						"cannot downgrade %q from %d to %d", name,
						old.Version, m.Version,
					)
				}
			}

			app, err := a.stubOrMake(name)
			if err != nil {
				return errcode.Annotatef(err, "get %q app stub", name)
			}

			action := "install"
			if old != nil { // This is an upgrade.
				action = "upgrade"
			}

			log.Printf("%s %s", action, name)
			if err := app.change(old, m); err != nil {
				return errcode.Annotatef(err, "%s %q", action, name)
			}

			a.state.setMeta(name, m)
			if err := a.saveState(); err != nil {
				return errcode.Annotatef(
					err, "save state after %s %q", action, name,
				)
			}
		}
	}

	toUninstall := uninstall.RevFloors()
	for _, floor := range toUninstall {
		for _, m := range floor {
			name := m.Name
			old := a.state.meta(name)
			if old == nil { // Should not happen.
				log.Printf("%q lost on uninstallaion", name)
				continue
			}

			app, err := a.stub(name)
			if err != nil {
				return errcode.Annotatef(err, "get %q app stub", name)
			}

			log.Printf("uninstall %s", name)
			if err := app.change(old, nil); err != nil {
				return errcode.Annotatef(err, "uninstall %q", name)
			}

			a.state.setMeta(name, nil)
			if err := a.saveState(); err != nil {
				return errcode.Annotatef(
					err, "save state after uninstall %q", name,
				)
			}
			a.removeStub(name)
		}
	}

	a.state.Anchored = strutil.MakeSet(anchored)
	if err := a.saveState(); err != nil {
		return errcode.Annotate(err, "save anchored list")
	}

	return nil
}

func (a *apps) install(names []string) error {
	anchored := make(map[string]bool)
	for name := range a.state.Anchored {
		anchored[name] = true
	}
	for _, name := range names {
		anchored[name] = true
	}
	return a.apply(strutil.SortedList(anchored))
}

func (a *apps) uninstall(names []string) error {
	anchored := make(map[string]bool)
	for name := range a.state.Anchored {
		anchored[name] = true
	}
	for _, name := range names {
		if !anchored[name] {
			return errcode.InvalidArgf("%q is not yet installed", name)
		}
		delete(anchored, name)
	}
	return a.apply(strutil.SortedList(anchored))
}

func (a *apps) update() error { return a.apply(a.anchored()) }

func (a *apps) anchored() []string {
	anchored := make(map[string]bool)
	for name := range a.state.Anchored {
		anchored[name] = true
	}
	return strutil.SortedList(anchored)
}

func (a *apps) list() []string { return a.state.list() }

func (a *apps) semVersions() map[string]string {
	return a.state.semVersions()
}
