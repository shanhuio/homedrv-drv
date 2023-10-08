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
	"testing"

	"encoding/json"
	"reflect"

	"shanhu.io/g/errcode"
	"shanhu.io/g/strutil"
	"shanhu.io/homedrv/drv/drvapi"
)

type simpleAppsStateStore struct {
	bs []byte
}

func (s *simpleAppsStateStore) save(state *appsState) error {
	bs, err := json.Marshal(state)
	if err != nil {
		return err
	}
	s.bs = bs
	return nil
}

func (s *simpleAppsStateStore) load() (*appsState, error) {
	state := new(appsState)
	if len(s.bs) == 0 {
		return state, nil
	}
	if err := json.Unmarshal(s.bs, state); err != nil {
		return nil, err
	}
	return state, nil
}

func newSimpleAppsStateStore() *simpleAppsStateStore {
	return &simpleAppsStateStore{}
}

type fakeAppState struct {
	version int64
	image   string
}

type fakeSystem struct {
	manifest map[string]*drvapi.AppMeta
	running  map[string]bool
	states   map[string]*fakeAppState
}

func newFakeSystem(m map[string]*drvapi.AppMeta) *fakeSystem {
	return &fakeSystem{
		manifest: m,
		running:  make(map[string]bool),
		states:   make(map[string]*fakeAppState),
	}
}

func (s *fakeSystem) makeStub(name string) (*appStub, error) {
	if _, ok := s.manifest[name]; !ok {
		return nil, errcode.NotFoundf("app %q not found", name)
	}
	return &appStub{App: &fakeApp{sys: s, name: name}}, nil
}

func (s *fakeSystem) meta(name string) (*drvapi.AppMeta, error) {
	m, ok := s.manifest[name]
	if !ok {
		return nil, errcode.NotFoundf("app %q not found", name)
	}
	return m, nil
}

type fakeApp struct {
	sys  *fakeSystem
	name string
}

func (a *fakeApp) Change(from, to *drvapi.AppMeta) error {
	a.Stop()
	if to == nil {
		delete(a.sys.states, a.name)
		return nil
	}

	for _, dep := range to.Deps {
		if a.sys.states[dep] == nil {
			return errcode.Internalf(
				"dep %q of %q not installed", dep, a.name,
			)
		}
	}
	a.sys.states[a.name] = &fakeAppState{
		version: to.Version,
		image:   to.Image,
	}
	return a.Start()
}

func (a *fakeApp) Start() error {
	a.sys.running[a.name] = true
	return nil
}

func (a *fakeApp) Stop() error {
	delete(a.sys.running, a.name)
	return nil
}

type testApps struct {
	*apps
	store *simpleAppsStateStore
	sys   *fakeSystem
}

func newTestApps(metas []*drvapi.AppMeta) (*testApps, error) {
	manifest := makeManifest(metas)
	store := newSimpleAppsStateStore()
	sys := newFakeSystem(manifest)
	config := &appsConfig{
		store:   store,
		querier: sys,
		maker:   sys,
	}
	apps, err := newApps(config)
	if err != nil {
		return nil, err
	}

	return &testApps{
		store: store,
		sys:   sys,
		apps:  apps,
	}, nil
}

func TestApps_empty(t *testing.T) {
	apps, err := newTestApps(nil)
	if err != nil {
		t.Fatal("create apps: ", err)
	}

	list := apps.list()
	if len(list) != 0 {
		t.Errorf("got %q, want empty list", list)
	}
}

func TestApps_single(t *testing.T) {
	apps, err := newTestApps([]*drvapi.AppMeta{{
		Name:    "killer",
		Version: 1,
	}})
	if err != nil {
		t.Fatal("create apps: ", err)
	}

	const appName = "killer"

	if err := apps.install([]string{appName}); err != nil {
		t.Fatal("install app")
	}
	if list := apps.list(); !reflect.DeepEqual(list, []string{appName}) {
		t.Fatalf("got %q, want single app %q", list, appName)
	}

	if !apps.sys.running[appName] {
		t.Error("app not running after installation")
	}
	if v := apps.sys.states[appName].version; v != 1 {
		t.Errorf("got version %d, want 1", v)
	}

	if err := apps.uninstall([]string{"killer"}); err != nil {
		t.Fatal("uninstall killer app")
	}

	if list := apps.list(); len(list) != 0 {
		t.Errorf("got %q after uninstall, want empty list", list)
	}

	if apps.sys.running[appName] {
		t.Error("app still running after uninstallation")
	}
	if len(apps.sys.running) != 0 {
		t.Errorf(
			"zombie app on system: %q",
			strutil.SortedList(apps.sys.running),
		)
	}
}

func TestApps_deps(t *testing.T) {
	apps, err := newTestApps([]*drvapi.AppMeta{{
		Name: "db", Version: 1,
	}, {
		Name: "memcached", Version: 2,
	}, {
		Name: "app1", Version: 3,
		Deps: []string{"db"},
	}, {
		Name: "app2", Version: 4,
		Deps: []string{"db", "memcached"},
	}})
	if err != nil {
		t.Fatal("create apps: ", err)
	}

	checkApps := func(when string, want []string) {
		t.Helper()

		if len(want) == 0 {
			if list := apps.list(); len(list) != 0 {
				t.Errorf("list %s, got %q, want empty", when, list)
			}
		} else {
			if list := apps.list(); !reflect.DeepEqual(list, want) {
				t.Errorf("list %s, got %q, want %q", when, list, want)
			}
		}

		for _, name := range want {
			if apps.sys.states[name] == nil {
				t.Errorf("version of %q is nil %s", name, when)
			}
			if !apps.sys.running[name] {
				t.Errorf("app %q is not running %s", name, when)
			}
		}
	}

	if err := apps.install([]string{"app1"}); err != nil {
		t.Fatal("install app1")
	}
	checkApps("after install app1", []string{"app1", "db"})

	if err := apps.install([]string{"app2"}); err != nil {
		t.Fatal("install app2")
	}
	checkApps("after install app2", []string{
		"app1", "app2", "db", "memcached",
	})

	if err := apps.uninstall([]string{"app1"}); err != nil {
		t.Fatal("unintall app1")
	}
	checkApps("after uninstall app1", []string{
		"app2", "db", "memcached",
	})

	if err := apps.uninstall([]string{"app2"}); err != nil {
		t.Fatal("uninstall app2")
	}
	checkApps("after uninstall app2", []string{})

	if len(apps.sys.running) != 0 {
		t.Errorf(
			"zombie app on system: %q",
			strutil.SortedList(apps.sys.running),
		)
	}
}

func TestApps_circleSelfDep(t *testing.T) {
	apps, err := newTestApps([]*drvapi.AppMeta{{
		Name: "snake", Deps: []string{"snake"},
	}})
	if err != nil {
		t.Fatal("create apps: ", err)
	}

	if err := apps.install([]string{"snake"}); err == nil {
		t.Fatal("install circular deps app got no error")
	} else {
		t.Logf("install circular dep: %s", err)
	}
}

func TestApps_circleDep(t *testing.T) {
	apps, err := newTestApps([]*drvapi.AppMeta{{
		Name: "x", Deps: []string{"y"},
	}, {
		Name: "y", Deps: []string{"z"},
	}, {
		Name: "z", Deps: []string{"x"},
	}, {
		Name: "app", Deps: []string{"x"},
	}})
	if err != nil {
		t.Fatal("create apps: ", err)
	}

	if err := apps.install([]string{"app"}); err == nil {
		t.Fatal("install circular deps app got no error")
	} else {
		t.Logf("install long circular dep: %s", err)
	}
}

func TestApps_notFound(t *testing.T) {
	apps, err := newTestApps([]*drvapi.AppMeta{{
		Name: "app", Deps: []string{"limbo"},
	}})
	if err != nil {
		t.Fatal("create apps: ", err)
	}

	if err := apps.install([]string{"app"}); !errcode.IsNotFound(err) {
		t.Fatalf("install invalid app got %q, want not found", err)
	} else {
		t.Logf("install invalid app: %s", err)
	}
}
