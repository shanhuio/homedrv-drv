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
	"strings"

	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
)

func signInRedirect(c *aries.C) error {
	// TODO(h8liu): add sign-in redirect URL.
	c.Redirect("/")
	return nil
}

func serveDashboard(s *server, c *aries.C) error {
	if c.Req.Method != "GET" {
		return errcode.InvalidArgf("request must be get")
	}

	aries.NeverCache(c)
	if c.User == "" {
		signInRedirect(c)
		return nil
	}

	d, err := newDashboardData(s, c, &DashboardDataRequest{
		Path: strings.TrimPrefix(c.Path, "/"),
	})
	if err != nil {
		return err
	}
	dat := struct{ Data *DashboardData }{Data: d}
	return s.tmpls.Serve(c, "dashboard.html", &dat)
}
