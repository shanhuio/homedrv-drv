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
	"strings"

	"shanhu.io/aries"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/timeutil"
)

// DashboardDataRequest is the AJAX request to load dashboard data.
type DashboardDataRequest struct {
	Path string
}

// DashboardData contains the page data for a particular dashboard
// state.
type DashboardData struct {
	Path     string
	Now      *timeutil.Timestamp // Unix seconds.
	NeedSudo bool                // Needs to get sudo cookie first.

	Overview      *DashboardOverviewData     `json:",omitempty"`
	TwoFactorAuth *Dashboard2FAData          `json:",omitempty"`
	SecurityLogs  *DashboardSecurityLogsData `json:",omitempty"`
	SSHKeys       *DashboardSSHKeysData      `json:",omitempty"`
}

func newDashboardData(s *server, c *aries.C, req *DashboardDataRequest) (
	*DashboardData, error,
) {
	d := &DashboardData{
		Path: req.Path,
		Now:  timeutil.TimestampNow(),
	}

	switch req.Path {
	case "2fa/enable-totp", "2fa/disable-totp":
		if err := s.sudoSessions.Check(c); err != nil {
			if !errcode.IsUnauthorized(err) {
				return nil, errcode.Annotate(err, "check sudo")
			}
			d.NeedSudo = true
			return d, nil
		}
	}

	switch req.Path {
	default:
		return nil, errcode.InvalidArgf("invalid path: %q", req.Path)
	case "overview":
		overview, err := newDashboardOverviewData(s)
		if err != nil {
			return nil, err
		}
		d.Overview = overview
	case "2fa", "2fa/enable-totp", "2fa/disable-totp":
		sub := strings.TrimPrefix(req.Path, "2fa/")
		twoFA, err := newDashboard2FAData(s, c, sub)
		if err != nil {
			return nil, err
		}
		d.TwoFactorAuth = twoFA
	case "change-password":
		// do nothing
	case "security-logs":
		dat, err := newDashboardSecurityLogsData(s, c)
		if err != nil {
			return nil, err
		}
		d.SecurityLogs = dat
	case "ssh-keys":
		dat, err := newDashboardSSHKeysData(s, c)
		if err != nil {
			return nil, err
		}
		d.SSHKeys = dat
	}
	return d, nil
}

func dashboardAPI(s *server) *aries.Router {
	dataHandler := func(c *aries.C, req *DashboardDataRequest) (
		*DashboardData, error,
	) {
		return newDashboardData(s, c, req)
	}

	r := aries.NewRouter()
	r.Call("data", dataHandler)
	return r
}
