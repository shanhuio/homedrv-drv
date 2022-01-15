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
	"crypto/sha256"
	"time"

	"shanhu.io/aries"
	"shanhu.io/aries/oauth2"
	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/osutil"
	"shanhu.io/misc/signer"
	"shanhu.io/pisces/settings"
)

type server struct {
	*backend
	drive       *drive
	appRegistry *appRegistry
	apps        *apps

	auth          *oauth2.Module
	sudoSessions  *sudoSessions
	loginSessions *loginSessions
	totp          *totp
	sshKeys       *sshKeys
	keyRegistry   *keyRegistry

	tmpls  *aries.Templates
	static *aries.StaticFiles

	updateSignal chan bool
}

func newServer(h *osutil.Home, c *drvcfg.Config) (*server, error) {
	back, err := newBackend("")
	if err != nil {
		return nil, errcode.Annotate(err, "create backend")
	}

	rel := new(drvapi.Release)
	if err := back.settings.Get(keyBuild, rel); err != nil {
		if !errcode.IsNotFound(err) {
			return nil, errcode.Annotate(err, "read current build")
		}
	}
	appReg := newAppRegistry(rel)
	stateStore := &appsStateSettings{
		key:      keyAppsState,
		settings: back.settings,
	}
	if rel.Name != "" { // not first time install.
		if err := maybeSetAppsStateFromLegacy(stateStore, appReg); err != nil {
			return nil, errcode.Annotate(err, "set apps state from legacy")
		}
	}

	apps, err := newApps(&appsConfig{
		store:   stateStore,
		querier: appReg,
	})
	if err != nil {
		return nil, errcode.Annotate(err, "build apps control")
	}

	objs, err := newObjects(h.Var("objs"))
	if err != nil {
		return nil, errcode.Annotate(err, "create objects store")
	}

	kernel := &kernel{
		settings:    back.settings,
		appDomains:  back.appDomains,
		appRegistry: appReg,
		apps:        apps,
		objects:     objs,
	}
	drive, err := newDrive(c, kernel)
	if err != nil {
		return nil, err
	}

	if err := apps.setMaker(newBuiltInApps(drive)); err != nil {
		return nil, errcode.Annotate(err, "setup builtin app stubs")
	}

	sessionKey, err := settings.String(back.settings, keySessionHMAC)
	if err != nil {
		return nil, errcode.Annotate(err, "read session key")
	}
	sudoSessions := newSudoSessions(sessionKey)
	loginSessions := newLoginSessions(sessionKey)
	keyRegistry := newKeyRegistry(back.users)
	auth := oauth2.NewModule(&oauth2.Config{
		SessionKey: []byte(sessionKey),
		PreSignOut: func(c *aries.C) error {
			sudoSessions.ClearCookie(c)
			return nil
		},
		KeyRegistry: keyRegistry,
	})

	signerKey := sha256.Sum256([]byte("state:" + sessionKey))
	stateSigner := signer.New(signerKey[:])

	totpCfg := &totpConfig{
		sudo:        sudoSessions,
		stateSigner: stateSigner,
		logs:        back.securityLogs,
		issuer: func() (string, error) {
			v, err := settings.String(back.settings, keyMainDomain)
			if errcode.IsNotFound(err) {
				return "unknown.homedrive.io", nil
			}
			return v, err
		},
		now: time.Now,
	}
	totp, err := newTOTP(back.users, totpCfg)
	if err != nil {
		return nil, errcode.Annotate(err, "create totp")
	}

	return &server{
		backend:     back,
		drive:       drive,
		appRegistry: appReg,
		apps:        apps,

		auth:          auth,
		sudoSessions:  sudoSessions,
		loginSessions: loginSessions,
		totp:          totp,
		sshKeys:       newSSHKeys(drive),
		keyRegistry:   keyRegistry,

		tmpls:  aries.NewTemplates(h.Lib("tmpl"), nil),
		static: aries.NewStaticFiles(h.Lib("static")),

		updateSignal: make(chan bool),
	}, nil
}

func (s *server) Drive() *drive { return s.drive }

func (s *server) f(f func(s *server, c *aries.C) error) aries.Func {
	return func(c *aries.C) error { return f(s, c) }
}
