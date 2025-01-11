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

// Package ncfront provides a reverse proxy that runs in front of nextcloud to
// handle several nextcloud docker issues.
package ncfront

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type front struct {
	proxy *httputil.ReverseProxy
}

func newFront(destHost string) *front {
	u := &url.URL{Scheme: "http", Host: destHost}
	p := httputil.NewSingleHostReverseProxy(u)

	p.ModifyResponse = func(resp *http.Response) error {
		// Per https://tinyurl.com/yvtxa7bu
		resp.Header.Set("Strict-Transport-Security", "max-age=15552000")

		return nil
	}

	return &front{proxy: p}
}

var redirects = map[string]string{
	"/.well-known/caldav":  "/remote.php/dav/",
	"/.well-known/carddav": "/remote.php/dav/",
}

func (f *front) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	re, found := redirects[r.URL.Path]
	if found {
		http.Redirect(w, r, re, http.StatusMovedPermanently)
		return
	}
	f.proxy.ServeHTTP(w, r)
}

// Main is the main entrance of the nextcloud front proxy.
func Main() {
	destHost := flag.String("nextcloud", "", "nextcloud's host")
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	if *destHost == "" {
		if v, ok := os.LookupEnv("NEXTCLOUD"); ok {
			*destHost = v
		} else {
			const defaultAddr = "nextcloud:80"
			*destHost = defaultAddr
		}
	}

	f := newFront(*destHost)
	log.Printf("listening on %s, forwarding to %s", *addr, *destHost)
	if err := http.ListenAndServe(*addr, f); err != nil {
		log.Fatal(err)
	}
}
