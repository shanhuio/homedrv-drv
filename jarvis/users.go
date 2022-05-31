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
	"crypto/rand"
	"log"
	"time"

	"shanhu.io/aries"
	"shanhu.io/misc/argon2"
	"shanhu.io/misc/errcode"
	"shanhu.io/pisces"
)

func hashPassword(password string) (*argon2.Password, error) {
	return argon2.NewPassword([]byte(password), rand.Reader)
}

type users struct {
	t *pisces.KV

	onChangePassword func(user string)
}

func newUsers(b *pisces.Tables) *users {
	return &users{t: b.NewKV("users")}
}

func (b *users) setOnChangePassword(f func(user string)) {
	b.onChangePassword = f
}

func (b *users) create(user, password string) error {
	hashed, err := hashPassword(password)
	if err != nil {
		return errcode.Annotate(err, "hash password")
	}
	info := &userInfo{
		Name:           user,
		Argon2Password: hashed,
	}
	return b.t.Add(user, info)
}

func (b *users) get(user string) (*userInfo, error) {
	info := new(userInfo)
	if err := b.t.Get(user, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (b *users) has(user string) (bool, error) { return b.t.Has(user) }
func (b *users) remove(user string) error      { return b.t.Remove(user) }

func (b *users) mutate(user string, f func(info *userInfo) error) error {
	info := new(userInfo)
	return b.t.Mutate(user, info, func(v interface{}) error {
		return f(v.(*userInfo))
	})
}

func (b *users) setPassword(user, password string, old *string) error {
	hashed, err := hashPassword(password)
	if err != nil {
		return errcode.Annotate(err, "hash password")
	}
	if err := b.mutate(user, func(info *userInfo) error {
		if old != nil {
			if err := checkUserPassword(info, *old); err != nil {
				return err
			}
		}
		info.BcryptPassword = nil // Clear old bcrypt version.
		info.Argon2Password = hashed
		return nil
	}); err != nil {
		return err
	}

	if b.onChangePassword != nil {
		b.onChangePassword(user)
	}
	return nil
}

func (b *users) checkPassword(user, password string) error {
	now := time.Now()
	if err := b.loginRateLimit(user, now); err != nil {
		if err == errTooManyFailures {
			return err
		}
		return errcode.Annotate(err, "login rate limit")
	}

	info := new(userInfo)
	if err := b.t.Get(user, info); err != nil {
		return err
	}
	if err := checkUserPassword(info, password); err != nil {
		if err == errWrongPassword {
			b.recordLoginFailure(user, now)
		}
		return err
	}
	b.recordLoginSuccess(user)
	return nil
}

func (b *users) totpInfo(user string) (*totpInfo, error) {
	info := new(userInfo)
	if err := b.t.Get(user, info); err != nil {
		return nil, errcode.Annotate(err, "get user info")
	}
	return userTOTPInfo(info), nil
}

type changePasswordRequest struct {
	OldPassword string
	NewPassword string
}

type changePasswordResponse struct {
	Error string
}

func (b *users) apiChangePassword(c *aries.C, req *changePasswordRequest) (
	*changePasswordResponse, error,
) {
	old := req.OldPassword
	if err := b.setPassword(c.User, req.NewPassword, &old); err != nil {
		if err == errWrongPassword {
			return &changePasswordResponse{
				Error: "Incorrect old password.",
			}, nil
		}
		return nil, err
	}
	return &changePasswordResponse{}, nil
}

func (b *users) disableTOTP(user string) error {
	return b.mutate(user, func(info *userInfo) error {
		info.TwoFactor = nil
		return nil
	})
}

// activateTOTP actually activates TOTP as a 2-Factor authentication
// method for a user.
func (b *users) activateTOTP(user string, secret string) error {
	return b.mutate(user, func(info *userInfo) error {
		if info.TwoFactor == nil {
			info.TwoFactor = new(twoFactorInfo)
		}
		info.TwoFactor.TOTP = &totpInfo{Secret: secret}
		return nil
	})
}

func (b *users) loginRateLimit(user string, now time.Time) error {
	info, err := b.get(user)
	if err != nil {
		return errcode.Annotate(err, "get user")
	}
	if info.RecentLoginFailures == nil {
		return nil
	}
	if info.RecentLoginFailures.count(now) < 5 {
		return nil
	}
	return errTooManyFailures
}

func (b *users) recordLoginFailure(user string, now time.Time) {
	if err := b.mutate(user, func(info *userInfo) error {
		if info.RecentLoginFailures == nil {
			info.RecentLoginFailures = new(recentFailures)
		}
		info.RecentLoginFailures.add(now)
		return nil
	}); err != nil {
		log.Printf("record login failure: %s", err)
	}
}

func (b *users) recordLoginSuccess(user string) {
	if err := b.mutate(user, func(info *userInfo) error {
		if info.RecentLoginFailures != nil {
			info.RecentLoginFailures.clear()
		}
		return nil
	}); err != nil {
		log.Printf("record login success: %s", err)
	}
}

func (b *users) api() *aries.Router {
	r := aries.NewRouter()
	r.Call("changepwd", b.apiChangePassword)
	return r
}
