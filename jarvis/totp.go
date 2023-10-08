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
	"bytes"
	"encoding/base64"
	"image/png"
	"log"
	"time"

	"github.com/pquerna/otp"
	totppkg "github.com/pquerna/otp/totp"
	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
	"shanhu.io/g/signer"
)

const (
	totpDigits    = otp.DigitsSix
	totpAlgorithm = otp.AlgorithmSHA256
)

// newTOTPKey creates a new TOTP key for user.
// It does NOT activate TOTP for authentication just yet.
// Returns the newly created key and error.
func newTOTPKey(user, issuer string) (*otp.Key, error) {
	// Refresh secret everytime TOTP is enabled.
	opts := totppkg.GenerateOpts{
		Issuer:      issuer,
		AccountName: user,
		Digits:      totpDigits,
		Algorithm:   totpAlgorithm,
	}
	key, err := totppkg.Generate(opts)
	if err != nil {
		return nil, errcode.Annotate(err, "generate TOTP key")
	}
	return key, nil
}

type sessionChecker interface {
	Check(c *aries.C) error
}

type totpConfig struct {
	sudo        sessionChecker
	stateSigner *signer.Signer
	logs        *securityLogs
	issuer      func() (string, error)
	now         func() time.Time
}

type totp struct {
	users       *users
	sudo        sessionChecker
	stateSigner *signer.Signer
	logs        *securityLogs
	issuer      func() (string, error)
	now         func() time.Time
}

func newTOTP(users *users, c *totpConfig) (*totp, error) {
	issuer := c.issuer
	if issuer == nil {
		issuer = func() (string, error) {
			return "shanhu.io", nil
		}
	}
	now := c.now
	if now == nil {
		now = time.Now
	}
	return &totp{
		users:       users,
		sudo:        c.sudo,
		stateSigner: c.stateSigner,
		logs:        c.logs,
		issuer:      issuer,
		now:         now,
	}, nil
}

func (t *totp) log(c *aries.C, event string) {
	if t.logs == nil {
		return
	}
	if err := t.logs.recordTwoFactorEvent(
		c.User, methodTOTP, event,
	); err != nil {
		log.Println(err)
	}
}

// DisableTOTPRequest is the request to disable 2-factor authentication..
type DisableTOTPRequest struct{}

// DisableTOTPResponse is the response to disable 2-factor authentication..
type DisableTOTPResponse struct{}

func (t *totp) apiDisable(c *aries.C, req *DisableTOTPRequest) (
	*DisableTOTPResponse, error,
) {
	if err := t.sudo.Check(c); err != nil {
		return nil, errcode.Annotate(err, "check sudo session")
	}
	if err := t.users.disableTOTP(c.User); err != nil {
		return nil, errcode.Annotate(err, "disable totp")
	}
	t.log(c, "disable")
	return new(DisableTOTPResponse), nil
}

func makeBase64QRCode(key *otp.Key) (string, error) {
	const width = 200
	const height = 200

	img, err := key.Image(width, height)
	if err != nil {
		return "", errcode.Annotate(err, "generate QR code image")
	}

	buffer := new(bytes.Buffer)
	buffer.WriteString("data:image/png;base64,")

	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	if err := png.Encode(encoder, img); err != nil {
		return "", errcode.Annotate(err, "encode image as base64 string")
	}
	encoder.Close()

	return buffer.String(), nil
}

// TOTPSetup contains data fields needed for setting up TOTP authentication.
type TOTPSetup struct {
	SignedSecret []byte
	QRCode       string
	URL          string
}

func (t *totp) setup(user string) (*TOTPSetup, error) {
	issuer, err := t.issuer()
	if err != nil {
		return nil, errcode.Annotate(err, "get issuer")
	}
	key, err := newTOTPKey(user, issuer)
	if err != nil {
		return nil, errcode.Annotate(err, "enable totp")
	}

	png, err := makeBase64QRCode(key)
	if err != nil {
		return nil, errcode.Annotate(err, "generate QR code")
	}

	// Sign secret before returning it to the user.
	signedSecret := t.stateSigner.Sign([]byte(key.Secret()))

	return &TOTPSetup{
		SignedSecret: signedSecret,
		QRCode:       png,
		URL:          key.URL(),
	}, nil
}

// SetupTOTPRequest is the request to create a signed TOTP setup.
type SetupTOTPRequest struct{}

func (t *totp) apiSetup(c *aries.C, _ *SetupTOTPRequest) (*TOTPSetup, error) {
	if err := t.sudo.Check(c); err != nil {
		return nil, errcode.Annotate(err, "check sudo session")
	}
	return t.setup(c.User)
}

// EnableTOTPRequest is the request to activate TOTP authentication..
type EnableTOTPRequest struct {
	SignedSecret []byte
	OTP          string
}

// EnableTOTPResponse is the response to activate TOTP authentication..
type EnableTOTPResponse struct {
	Error string // Expected error that user should see.
}

func (t *totp) apiEnable(c *aries.C, req *EnableTOTPRequest) (
	*EnableTOTPResponse, error,
) {
	if err := t.sudo.Check(c); err != nil {
		return nil, errcode.Annotate(err, "check sudo session")
	}
	ok, secretBytes := t.stateSigner.Check(req.SignedSecret)
	if !ok {
		return nil, errcode.InvalidArgf("invalid secret signature")
	}
	secret := string(secretBytes)

	opts := totppkg.ValidateOpts{
		Digits:    totpDigits,
		Algorithm: totpAlgorithm,
	}
	if ok, err := totppkg.ValidateCustom(
		req.OTP, secret, t.now(), opts,
	); !ok || err != nil {
		return &EnableTOTPResponse{
			Error: "OTP code incorrect.",
		}, nil
	}

	user := c.User
	if err := t.users.activateTOTP(user, secret); err != nil {
		return nil, errcode.Annotate(err, "activate totp")
	}
	t.log(c, "enable")
	return &EnableTOTPResponse{}, nil
}

func (t *totp) api() *aries.Router {
	r := aries.NewRouter()
	r.Call("setup", t.apiSetup)
	r.Call("disable", t.apiDisable)
	r.Call("enable", t.apiEnable)
	return r
}

func totpValidate(passcode, secret string) (bool, error) {
	opts := totppkg.ValidateOpts{
		Digits:    totpDigits,
		Algorithm: totpAlgorithm,
	}
	t := time.Now()
	ok, err := totppkg.ValidateCustom(passcode, secret, t, opts)
	return ok, err
}
