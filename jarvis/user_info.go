// Copyright (C) 2023  Shanhu Tech Inc.
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
	"golang.org/x/crypto/bcrypt"
	"shanhu.io/g/argon2"
	"shanhu.io/g/errcode"
)

type userInfo struct {
	Name string

	BcryptPassword []byte           `json:",omitempty"`
	Argon2Password *argon2.Password `json:",omitempty"`
	TwoFactor      *twoFactorInfo   `json:",omitempty"`

	RecentLoginFailures *recentFailures `json:",omitempty"`
}

var errWrongPassword = errcode.Unauthorizedf("wrong password")

func checkUserPassword(info *userInfo, password string) error {
	if info.Argon2Password != nil {
		if !info.Argon2Password.Check([]byte(password)) {
			return errWrongPassword
		}
		return nil
	}

	// Fallback to old bcrypt checking.
	if err := bcrypt.CompareHashAndPassword(
		info.BcryptPassword, []byte(password),
	); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return errWrongPassword
		}
		return err
	}
	return nil
}

func userTOTPInfo(info *userInfo) *totpInfo {
	if info.TwoFactor == nil {
		return nil
	}
	return info.TwoFactor.TOTP
}
