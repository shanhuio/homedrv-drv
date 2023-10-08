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
	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
)

// DashboardTOTPData contains data for TOTP 2FA method.
type DashboardTOTPData struct {
	Enabled   bool
	TOTPSetup *TOTPSetup `json:",omitempty"`
}

// Dashboard2FAData contains data for 2-factor authentication tab.
type Dashboard2FAData struct {
	TOTP *DashboardTOTPData `json:",omitempty"`
}

func newDashboard2FAData(s *server, c *aries.C, sub string) (
	*Dashboard2FAData, error,
) {
	info, err := s.users.get(c.User)
	if err != nil {
		return nil, err
	}

	data := new(Dashboard2FAData)

	var totp *totpInfo
	if info.TwoFactor != nil && info.TwoFactor.TOTP != nil {
		totp = info.TwoFactor.TOTP
	}

	enabled := totp != nil
	data.TOTP = &DashboardTOTPData{Enabled: enabled}

	if sub == "enable-totp" && !enabled {
		setup, err := s.totp.setup(c.User)
		if err != nil {
			return nil, errcode.Annotate(err, "enable totp")
		}

		// We know it must be in disabled state at this point.
		data.TOTP.TOTPSetup = setup
	}

	return data, nil
}
