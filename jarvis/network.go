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
	"shanhu.io/virgo/dock"
)

func networkCIDRs(d *drive) ([]string, error) {
	info, err := dock.InspectNetwork(d.dock, d.network())
	if err != nil {
		return nil, err
	}
	if info.IPAM == nil {
		return nil, nil
	}
	var cidrs []string
	for _, c := range info.IPAM.Config {
		cidrs = append(cidrs, c.Subnet)
	}
	return cidrs, nil
}
