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

package homedial

import (
	"context"
	"net"
	"strings"
	"time"
)

var fallbackNetDialer = &net.Dialer{
	Timeout:   10 * time.Second,
	KeepAlive: 30 * time.Second,
}

var homedrvIPv4 = map[string]string{
	"homedrive.io":             "167.172.10.171",
	"www.homedrive.io":         "167.172.10.171",
	"fabrics.homedrive.io":     "178.128.130.77",
	"fabrics-ge.homedrive.io":  "157.245.24.167",
	"fabrics-ge1.homedrive.io": "206.81.25.26",
	"fabrics-sgp.homedrive.io": "149.28.152.149",
}

// Dial dials HomeDrive servers.
func Dial(ctx context.Context, network, addr string) (
	net.Conn, error,
) {
	if network == "tcp" || network == "tcp4" {
		// Manually resolve IPv4 addresses for fabrics. This by passes DNS
		// resolvers in user's home networks, which might be faulty.
		trimmed := strings.TrimSuffix(addr, ".")
		if ip, ok := homedrvIPv4[trimmed]; ok {
			addr = ip // Directly resolve to IP address.
		}
	}
	return fallbackNetDialer.DialContext(ctx, network, addr)
}
