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

	"shanhu.io/aries"
	"shanhu.io/aries/identity"
	"shanhu.io/misc/errcode"
)

func guestRouter(s *server) *aries.Router {
	serveStatic := s.static.Serve

	r := aries.NewRouter()

	r.Index(func(c *aries.C) error { return serveIndex(s, c) })

	for _, tab := range dashboardTabs {
		capture := tab.name
		r.Dir(tab.name, func(c *aries.C) error {
			return serveDashboard(s, c, capture)
		})
	}

	r.File("login", func(c *aries.C) error { return serveLogin(s, c) })
	r.File("confirm-password", func(c *aries.C) error {
		return serveConfirmPassword(s, c)
	})
	r.File("sudo", func(c *aries.C) error { return serveSudo(s, c) })
	r.File("input-totp", func(c *aries.C) error {
		return serveInputTOTP(s, c)
	})
	r.File("totp", func(c *aries.C) error { return serveCheckTOTP(s, c) })

	r.Get("style.css", serveStatic)
	r.Get("favicon.ico", serveStatic)
	r.Dir("js", serveStatic)
	r.Dir("jslib", serveStatic)
	r.Dir("img", serveStatic)
	r.Dir("fonts", serveStatic)

	return r
}

func userRouter(s *server) *aries.Router {
	r := aries.NewRouter()
	r.DirService("api", apiRouter(s))
	r.DirService("obj", s.drive.objects)
	return r
}

func apiRouter(s *server) *aries.Router {
	r := aries.NewRouter()
	r.Call("hello", func(c *aries.C, msg string) (string, error) {
		return msg, nil
	})
	r.DirService("user", s.users.api())
	r.DirService("totp", s.totp.api())
	r.DirService("sshkeys", s.sshKeys.api())
	r.DirService("dashboard", dashboardAPI(s))
	r.DirService("id", identity.NewService(s.identity))
	r.DirService("obj", s.drive.objects.api())

	// just stubbing
	r.Call("sys/push-update", func(c *aries.C, bs []byte) error {
		return pushManualUpdate(s.drive, bs)
	})
	return r
}

func adminRouter(s *server) *aries.Router {
	r := aries.NewRouter()
	r.DirService("api", adminAPIRouter(s))
	return r
}

func adminAPIRouter(s *server) *aries.Router {
	r := aries.NewRouter()
	r.Call("update", func(c *aries.C, sig bool) error {
		s.updateSignal <- sig
		return nil
	})
	r.Call("recreate-doorway", func(c *aries.C) error {
		go func() {
			if err := recreateDoorway(s.drive); err != nil {
				log.Println(errcode.Annotate(err, "recreate doorway"))
			}
		}()
		return nil
	})
	r.Call("set-password", func(c *aries.C, req *changePasswordRequest) error {
		return s.users.setPassword(rootUser, req.NewPassword)
	})
	r.Call("set-api-key", s.keyRegistry.apiSet)

	return r
}

func localRouter(s *server) *aries.Router {
	r := aries.NewRouter()
	r.Index(aries.StringFunc("welcome"))
	return r
}
