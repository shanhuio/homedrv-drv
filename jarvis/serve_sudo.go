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
	"net/url"

	"shanhu.io/pub/aries"
	"shanhu.io/pub/errcode"
)

func serveSudo(s *server, c *aries.C) error {
	if c.User == "" {
		return errcode.Unauthorizedf("user has not signed in")
	}

	if err := parsePostForm(c); err != nil {
		return err
	}

	pass := c.Req.PostFormValue("password")
	redirect := c.Req.PostFormValue("redirect")
	const user = rootUser
	if err := s.users.checkPassword(user, pass); err != nil {
		if errcode.IsUnauthorized(err) {
			c.Redirect(confirmPasswordURL(redirect, "wrong-password"))
			return nil
		}
		return aries.AltInternal(err, "failed to check password")
	}

	s.sudoSessions.SetCookie(c)
	c.Redirect(redirect)

	return nil
}

var confirmPasswordRedirectPaths = map[string]bool{
	"/2fa/enable-totp":  true,
	"/2fa/disable-totp": true,
}

func confirmPasswordURL(redirect, errMsg string) string {
	u := &url.URL{Path: "/confirm-password"}
	q := u.Query()
	q.Set("redirect", redirect)
	if errMsg != "" {
		q.Set("err", errMsg)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func serveConfirmPassword(s *server, c *aries.C) error {
	if c.User == "" {
		signInRedirect(c)
		return nil
	}

	type pageData struct {
		RedirectTo string
		Error      string
	}

	d := new(pageData)
	q := c.Req.URL.Query()
	r := q.Get("redirect")
	if r == "" {
		return errcode.InvalidArgf("redirection path missing")
	}
	if !confirmPasswordRedirectPaths[r] {
		return errcode.InvalidArgf("redirect path not allowed")
	}

	if q.Get("err") == "wrong-password" {
		d.Error = "Wrong password."
	}
	d.RedirectTo = r

	dat := struct{ Data *pageData }{Data: d}
	return s.tmpls.Serve(c, "confirmpass.html", &dat)
}
