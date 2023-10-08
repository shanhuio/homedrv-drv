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
	"net/url"

	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
	"shanhu.io/homedrv/drv/homeapp/nextcloud"
)

func serveLogin(s *server, c *aries.C) error {
	if err := parsePostForm(c); err != nil {
		return err
	}
	pass := c.Req.PostFormValue("password")
	const user = rootUser
	remoteIP := aries.RemoteIPString(c)
	if err := s.users.checkPassword(user, pass); err != nil {
		if errcode.IsUnauthorized(err) {
			if err != errTooManyFailures {
				if err := s.securityLogs.recordFailedLogin(
					user, remoteIP, "",
				); err != nil {
					log.Println(err)
				}
			}

			if err == errTooManyFailures {
				c.Redirect("/?err=too-many-failures")
			} else {
				c.Redirect("/?err=wrong-password")
			}
			return nil
		}
		return aries.AltInternal(err, "failed to check password")
	}

	// TOTP is the only supported OTP method today.
	totpInfo, err := s.users.totpInfo(user)
	if err != nil {
		return errcode.Annotate(err, "get TOTP config")
	}
	if totpInfo == nil {
		// TOTP not enabled. Directly set login cookie and redirect to /.
		if err := s.securityLogs.recordLogin(
			user, remoteIP, "",
		); err != nil {
			log.Println(err)
		}
		s.auth.SetupCookie(c, user)
		c.Redirect("/")
	} else {
		// TOTP enabled. Redirect to /input-totp with proper token.
		u := &url.URL{Path: "/input-totp"}
		q := u.Query()
		q.Set("token", s.loginSessions.Token(user))
		u.RawQuery = q.Encode()
		c.Redirect(u.String())
	}
	return nil
}

func serveCheckTOTP(s *server, c *aries.C) error {
	// Parse OTP out of request form.
	if c.Req.Method != "POST" {
		return errcode.InvalidArgf("request must be post")
	}
	if err := c.Req.ParseForm(); err != nil {
		return errcode.InvalidArgf("error parsing form: %v", err)
	}

	token := c.Req.PostFormValue("token")
	user, err := s.loginSessions.Check(token)
	if err != nil {
		return errcode.InvalidArgf("invalid TOTP session")
	}

	// Get TOTP secret.
	totpInfo, err := s.users.totpInfo(user)
	if err != nil {
		return aries.AltInternal(err, "check user TOTP")
	}
	if totpInfo == nil {
		return errcode.Internalf("missing TOTP config")
	}

	totp := c.Req.PostFormValue("totp")
	remoteIP := aries.RemoteIPString(c)

	ok, err := totpValidate(totp, totpInfo.Secret)
	if !ok || err != nil {
		if err := s.securityLogs.recordFailedLogin(
			user, remoteIP, "totp",
		); err != nil {
			log.Println(err)
		}

		u := &url.URL{Path: "/input-totp"}
		q := u.Query()
		q.Set("token", token)
		q.Set("err", "wrong-totp")
		u.RawQuery = q.Encode()

		c.Redirect(u.String())
		return nil
	}

	if err := s.securityLogs.recordLogin(
		user, remoteIP, "totp",
	); err != nil {
		log.Println(err)
	}

	s.auth.SetupCookie(c, user)
	c.Redirect("/")
	return nil
}

func serveInputTOTP(s *server, c *aries.C) error {
	q := c.Req.URL.Query()
	token := q.Get("token")
	if _, err := s.loginSessions.Check(token); err != nil {
		// Ask user to start sign-in flow again.
		signInRedirect(c)
		return nil
	}

	type pageData struct {
		SessionToken string
		Issuer       string
		LoginError   string
	}

	issuer, err := s.totp.issuer()
	if err != nil {
		return errcode.Annotate(err, "get issuer")
	}

	d := &pageData{
		SessionToken: token,
		Issuer:       issuer,
	}
	if q.Get("err") == "wrong-totp" {
		d.LoginError = "Wrong TOTP."
	}

	dat := struct{ Data *pageData }{Data: d}
	return s.tmpls.Serve(c, "inputtotp.html", &dat)
}

func serveCover(s *server, c *aries.C) error {
	type pageData struct {
		HideLogin  bool
		RedirectTo string
		LoginError string
	}

	d := new(pageData)

	var ncDomains []string
	if err := s.drive.settings.Get(
		nextcloud.KeyDomains, &ncDomains,
	); err != nil {
		if !errcode.IsNotFound(err) {
			return aries.AltInternal(err, "failed to load server config")
		}
	}
	if len(ncDomains) > 0 {
		d.RedirectTo = (&url.URL{
			Scheme: "https",
			Host:   ncDomains[0],
		}).String()
	}

	q := c.Req.URL.Query()
	switch q.Get("err") {
	case "wrong-password":
		d.LoginError = "Wrong password."
	case "too-many-failures":
		d.LoginError = "Too many login failures recently."
	}

	dat := struct{ Data *pageData }{Data: d}
	return s.tmpls.Serve(c, "cover.html", &dat)
}

func serveIndex(s *server, c *aries.C) error {
	if c.User == "" {
		return serveCover(s, c)
	}
	c.Redirect("/overview")
	return nil
}
