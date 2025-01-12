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

// Package doorway is the HTTP frontend on a shanhu instance.
package doorway

import (
	"context"
	"flag"
	"log"
)

// Main is the main entrance for doorway binary.
func Main() {
	var (
		httpsAddr = flag.String("https", ":8443", "HTTPS address to listen on.")
		httpAddr  = flag.String("http", ":8080", "HTTP address to listen on.")
		home      = flag.String("home", ".", "home directory")
	)
	flag.Parse()

	ctx := context.Background()

	config, err := ConfigFromHome(*home)
	if err != nil {
		log.Fatal(err)
	}

	config.LocalAddr = *httpsAddr
	if *httpAddr != "" {
		config.HTTPServer.Addr = *httpAddr
	}

	if err := Serve(ctx, config); err != nil {
		log.Fatal(err)
	}
}
