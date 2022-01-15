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

package doorway

import (
	"net/http"
	"net/url"
)

func forwardToHTTP(req *http.Request, host string) {
	req.Header.Set("X-Forwarded-Host", req.Host)

	// The remote address that doorway sees is always the real address, as it
	// is either directly listening on the https port, or it gets the remote
	// address from the fabrics tunnel. This avoids IP spoofing that might
	// confuse Nextcloud or other internal applications.
	req.Header.Del("X-Forwarded-For")
	req.Header.Del("X-Real-IP")

	if len(req.RemoteAddr) > 0 && req.RemoteAddr[0] == '|' {
		req.RemoteAddr = req.RemoteAddr[1:] // trim the '|'
		req.Header.Add("Via", "1.0 hometunn")
	}

	u := req.URL      // Modify the URL.
	u.Scheme = "http" // Terminates http.
	u.Host = host
}

var sinkURL = &url.URL{
	Scheme: "http",
	Host:   "localhost",
}
