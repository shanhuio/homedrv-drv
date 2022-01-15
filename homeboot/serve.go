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

package homeboot

import (
	"log"
	"net"
	"time"

	"github.com/hashicorp/mdns"
	"shanhu.io/aries"
	"shanhu.io/misc/errcode"
)

func serveMDNS(ip net.IP) error {
	service, err := mdns.NewMDNSService(
		"homedrive",
		"_http._tcp.",
		"local.",
		"homedrive.local.", // hostname
		80,
		[]net.IP{ip},
		[]string{"HomeDrive, www.homedrive.io"},
	)
	if err != nil {
		return errcode.Annotate(err, "make mDNS service")
	}

	config := &mdns.Config{Zone: service}
	s, err := mdns.NewServer(config)
	if err != nil {
		return errcode.Annotate(err, "make mDNS server")
	}
	log.Println("mdns server started")
	defer s.Shutdown()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
	}
	return nil
}

func cmdServe(args []string) error {
	flags := cmdFlags.New()
	useMDNS := flags.Bool("mdns", true, "serve mDNS")
	ip := flags.String("ip", "", "ip address to serve on")
	addr := flags.String("addr", ":8080", "address to serve the service")
	flags.ParseArgs(args)

	if *ip == "" {
		return errcode.InvalidArgf("ip address not specified")
	}

	if *useMDNS {
		parsedIP := net.ParseIP(*ip)
		if parsedIP == nil {
			return errcode.InvalidArgf("invalid ip: %q", *ip)
		}

		go func() {
			if err := serveMDNS(parsedIP); err != nil {
				log.Fatal(err)
			}
		}()
	}

	server := aries.StringFunc("homedrive installer")
	log.Printf("serving at %s", *addr)
	return aries.ListenAndServe(*addr, server)
}
