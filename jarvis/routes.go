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
	"shanhu.io/aries"
	"shanhu.io/aries/identity"
)

func guestRouter(s *server) *aries.Router {
	r := aries.NewRouter()

	r.Index(func(c *aries.C) error { return serveIndex(s, c) })

	dash := s.f(serveDashboard)
	r.Get("overview", dash)
	r.Get("ssh-keys", dash)
	r.Get("security-logs", dash)
	r.Get("change-password", dash)
	r.Get("2fa", dash)
	r.Get("2fa/enable-totp", dash)
	r.Get("2fa/disable-totp", dash)

	r.File("login", s.f(serveLogin))
	r.File("confirm-password", s.f(serveConfirmPassword))
	r.File("sudo", s.f(serveSudo))
	r.File("input-totp", s.f(serveInputTOTP))
	r.File("totp", s.f(serveCheckTOTP))

	static := s.static.Serve
	r.Get("style.css", static)
	r.Get("favicon.ico", static)
	r.Dir("js", static)
	r.Dir("jslib", static)
	r.Dir("img", static)
	r.Dir("fonts", static)

	return r
}

func userRouter(s *server, api aries.Service) *aries.Router {
	r := aries.NewRouter()
	r.DirService("api", api)
	r.DirService("obj", s.drive.objects)
	return r
}

func apiRouter(s *server) *aries.Router {
	r := aries.NewRouter()
	r.DirService("user", s.users.api())
	r.DirService("totp", s.totp.api())
	r.DirService("sshkeys", s.sshKeys.api())
	r.DirService("dashboard", dashboardAPI(s))
	r.DirService("id", identity.NewService(s.identity))
	r.DirService("obj", s.drive.objects.api())
	r.DirService("admin", adminTasksAPI(s))
	return r
}
