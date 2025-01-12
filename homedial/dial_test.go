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

package homedial

import (
	"testing"
)

func TestMapAddress(t *testing.T) {
	for _, test := range []struct {
		net, addr, addrWant string
	}{{
		net:      "tcp",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "178.128.130.77:443",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "178.128.130.77:443",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io:80",
		addrWant: "178.128.130.77:80",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io.:80",
		addrWant: "178.128.130.77:80",
	}, {
		net:      "tcp6",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "fabrics.homedrive.io:443",
	}, {
		net:      "udp",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "fabrics.homedrive.io:443",
	}} {
		got := mapAddress(test.net, test.addr)
		if got != test.addrWant {
			t.Errorf(
				"map net=%s addr=%q, got %q, want %q",
				test.net, test.addr, test.addrWant, got,
			)
		}
	}
}
