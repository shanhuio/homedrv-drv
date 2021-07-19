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
)

// DashboardSecurityLogsData encapsulates security logs entries
// for the dashbaord.
type DashboardSecurityLogsData struct {
	Entries []*LogEntry
}

func newDashboardSecurityLogsData(s *server, c *aries.C) (
	*DashboardSecurityLogsData, error,
) {
	// TODO(h8liu): add pages
	entries, err := s.securityLogs.list(0)
	if err != nil {
		return nil, aries.AltInternal(err, "fail to fetch security logs")
	}

	for _, entry := range entries {
		entry.TSec = time.Unix(0, entry.T).Unix()
	}

	return &DashboardSecurityLogsData{
		Entries: entries,
	}, nil
}
