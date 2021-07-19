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
	"time"

	"shanhu.io/aries"
	"shanhu.io/misc/errcode"
)

const (
	tabOverview       = "overview"
	tabChangePassword = "change-password"
	tabTwoFactorAuth  = "2fa"
	tabSecurityLogs   = "security-logs"
	tabSSHKeys        = "ssh-keys"
)

type dashboardTab struct {
	name string
	subs []string
}

var dashboardTabs = []dashboardTab{
	{name: tabOverview},
	{name: tabChangePassword},
	{name: tabTwoFactorAuth, subs: []string{
		"enable-totp",
		"disable-totp",
	}},
	{name: tabSecurityLogs},
	{name: tabSSHKeys},
}

func checkDashboardSub(sub string, allowedSubs []string) error {
	// Root subs are always allowed. E.g., /overview, /2fa, etc.
	if sub == "" {
		return nil
	}
	// Otherwise, we need to check if this sub is allowed for the tab.
	for _, s := range allowedSubs {
		if s == sub {
			return nil
		}
	}
	return errcode.InvalidArgf("invalid path: %q", sub)
}

func checkDashboardTab(tab, sub string) error {
	for _, t := range dashboardTabs {
		if t.name != tab {
			continue
		}
		if err := checkDashboardSub(sub, t.subs); err == nil {
			return nil
		}
	}
	return errcode.InvalidArgf("invalid state: %q or path: %q", tab, sub)
}

// DashboardDataRequest is the AJAX request to load dashboard data.
type DashboardDataRequest struct {
	Tab       string
	Sub       string
	RequestID int
}

// DashboardData contains the page data for a particular dashboard
// state.
type DashboardData struct {
	Tab       string
	Sub       string
	RequestID int

	Now      int64 // Unix seconds.
	NeedSudo bool  // Needs to get sudo cookie first.

	Overview      *DashboardOverviewData     `json:",omitempty"`
	TwoFactorAuth *Dashboard2FAData          `json:",omitempty"`
	SecurityLogs  *DashboardSecurityLogsData `json:",omitempty"`
	SSHKeys       *DashboardSSHKeysData      `json:",omitempty"`
}

func newDashboardData(s *server, c *aries.C, req *DashboardDataRequest) (
	*DashboardData, error,
) {
	if err := checkDashboardTab(req.Tab, req.Sub); err != nil {
		return nil, err
	}
	d := &DashboardData{
		Tab:       req.Tab,
		Sub:       req.Sub,
		RequestID: req.RequestID,
		Now:       time.Now().Unix(),
	}

	if d.Tab == tabTwoFactorAuth &&
		(req.Sub == "enable-totp" || req.Sub == "disable-totp") {
		if err := s.sudoSessions.Check(c); err != nil {
			if !errcode.IsUnauthorized(err) {
				return nil, errcode.Annotate(err, "check sudo")
			}
			d.NeedSudo = true
			return d, nil
		}
	}

	switch d.Tab {
	case tabOverview:
		overview, err := newDashboardOverviewData(s)
		if err != nil {
			return nil, err
		}
		d.Overview = overview
	case tabTwoFactorAuth:
		twoFA, err := newDashboard2FAData(s, c, req.Sub)
		if err != nil {
			return nil, err
		}
		d.TwoFactorAuth = twoFA
	case tabSecurityLogs:
		dat, err := newDashboardSecurityLogsData(s, c)
		if err != nil {
			return nil, err
		}
		d.SecurityLogs = dat
	case tabSSHKeys:
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
