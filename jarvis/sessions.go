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
	"time"

	"shanhu.io/aries"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/hashutil"
	"shanhu.io/misc/signer"
)

const sudoSessionsCookie = "sudo"

type sudoSessions struct {
	s *signer.Sessions
}

func newSudoSessions(sessionKey string) *sudoSessions {
	// Derive sudoSessions key from sessionKey.
	sudoSessionsKey := hashutil.HashStr("sudo:" + sessionKey)

	const sudoSessionsValidDuration = 10 * time.Minute
	s := signer.NewSessions([]byte(sudoSessionsKey), sudoSessionsValidDuration)

	return &sudoSessions{s: s}
}

func (s *sudoSessions) Check(c *aries.C) error {
	if c.User == "" {
		return errcode.Unauthorizedf("user not signed in")
	}

	cookie := c.ReadCookie(sudoSessionsCookie)
	content, _, ok := s.s.Check(cookie)
	if !ok {
		return errcode.Unauthorizedf("cookie has expired")
	}
	if string(content) != c.User {
		return errcode.Unauthorizedf("cookie user invalid")
	}
	return nil
}

func (s *sudoSessions) SetCookie(c *aries.C) {
	// Save username as the content of the sudo cookie.
	token, expires := s.s.New([]byte(c.User), 0)
	c.WriteCookie(sudoSessionsCookie, token, expires)
}

func (s *sudoSessions) ClearCookie(c *aries.C) {
	c.ClearCookie(sudoSessionsCookie)
}

type loginSessions struct {
	s *signer.Sessions
}

func newLoginSessions(key string) *loginSessions {
	// Derive 2f session key from base sessionKey.
	loginSessionsKey := hashutil.HashStr("2factor:" + key)

	const loginSessionsDuration = 3 * time.Minute
	s := signer.NewSessions([]byte(loginSessionsKey), loginSessionsDuration)

	return &loginSessions{
		s: s,
	}
}

// Returns decoded username and nil if token is valid.
func (s *loginSessions) Check(token string) (string, error) {
	b, _, ok := s.s.Check(token)
	if !ok {
		return "", errcode.Unauthorizedf("can not verify 2factor session")
	}
	return string(b), nil
}

func (s *loginSessions) Token(user string) string {
	token, _ := s.s.New([]byte(user), 0 /* default ttl */)
	return token
}
