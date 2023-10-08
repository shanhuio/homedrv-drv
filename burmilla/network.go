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

package burmilla

import (
	"strings"

	"shanhu.io/g/errcode"
)

// HostIPs returns the ip addresses of the given network device.
func HostIPs(b *Burmilla, dev string) ([]string, error) {
	args := strings.Fields(
		"ip -br -family inet address show dev",
	)
	args = append(args, dev)
	out, err := b.ExecOutput(args)
	if err != nil {
		return nil, err
	}

	// Output is in form of:
	//   eth0    UP    x.x.x.x/xx x.x.x.x/xx ...
	// fields[2:] should all be IPv4 addresses
	s := string(out)
	fields := strings.Fields(s)
	if len(fields) < 3 {
		return nil, errcode.Internalf("unexpected output: %q", s)
	}

	ips := append([]string{}, fields[2:]...)
	return ips, nil
}
