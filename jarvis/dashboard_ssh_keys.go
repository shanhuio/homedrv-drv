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
	"strings"

	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
)

// DashboardSSHKeysData contains data for initializing SSH Keys dashboard page.
type DashboardSSHKeysData struct {
	Keys     string
	Disabled bool
}

func newDashboardSSHKeysData(s *server, _ *aries.C) (
	*DashboardSSHKeysData, error,
) {
	if !s.drive.hasSys() {
		return &DashboardSSHKeysData{Disabled: true}, nil
	}

	keys, err := s.sshKeys.list()
	if err != nil {
		return nil, errcode.Annotate(err, "get ssh keys")
	}

	dat := new(DashboardSSHKeysData)
	if len(keys) > 0 {
		dat.Keys = strings.Join(keys, "\n") + "\n"
	}

	return dat, nil
}
