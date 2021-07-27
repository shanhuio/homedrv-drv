package jarvis

import (
	"golang.org/x/crypto/bcrypt"
	"shanhu.io/misc/errcode"
)

const rootUser = "root"

type userInfo struct {
	Name string

	BcryptPassword []byte
	TwoFactor      *twoFactorInfo `json:",omitempty"`

	RecentLoginFailures *recentFailures `json:",omitempty"`

	APIKeys []byte
}

func bcryptPassword(pw string) ([]byte, error) {
	// TODO(h8liu): use argon2
	return bcrypt.GenerateFromPassword([]byte(pw), 0)
}

var errWrongPassword = errcode.Unauthorizedf("wrong password")

func checkUserPassword(info *userInfo, password string) error {
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
