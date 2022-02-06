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

package jarvis

import (
	"flag"
	"log"

	"shanhu.io/aries"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/homedrv/homeboot"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/osutil"

	_ "github.com/lib/pq"  // for postgres
	_ "modernc.org/sqlite" // sqlite db driver
)

// Main is the main entrance of jarvis server or client program.
func Main() {
	if osutil.Arg0Base() == "jarvisd" {
		serverMain()
		return
	}
	clientMain()
}

func makeService(s *server, admin aries.Service) aries.Service {
	return &aries.ServiceSet{
		Auth:  s.auth.Auth(),
		User:  userRouter(s),
		Admin: admin,
		Guest: guestRouter(s),
	}
}

func bg(s *server) {
	d := s.Drive()

	// Before starting the system tasks scheduler, make sure the system is
	// properlly installed.
	installed, err := d.settings.Has(keyBuild)
	if err != nil {
		// Basic install check failed.
		log.Println("check installed: ", err)
	} else if !installed { // This is first time.
		if err := downloadAndInstall(d); err != nil {
			log.Println("install failed: ", err)
		}
	} else { // Not first time.
		if err := maybeFinishUpdate(d); err != nil {
			log.Println("update failed: ", err)
			// It is important to proceed here, as the next update might be
			// able to fix the issue. At this point, the apps are in
			// undefiend state, but jarvis is already on the latest.
		}
		fixThings(d)
	}

	// Start the background update heartbeat querier.
	if d.config.Bare {
		log.Println("running in bare mode, no update in background")
	} else if d.config.Channel != "" {
		// Subscribe channel and maybe schedule update task.
		go cronUpdateOnChannel(d, s.updateSignal)
	}

	go cronNextcloud(d)

	d.tasks.bg() // Handle background system tasks now.
}

func run(homeDir, addr string) error {
	h, err := osutil.NewHome(homeDir)
	if err != nil {
		return errcode.Annotate(err, "open home dir")
	}

	// jarvis reads config from var.
	config, err := readConfig(h.Var("config.jsonx"))
	if err != nil {
		return errcode.Annotate(err, "read config")
	}

	s, err := newServer(h, config)
	if err != nil {
		return errcode.Annotate(err, "create server")
	}
	empty := drvcfg.Image(config.Naming, "empty")
	if err := homeboot.BuildEmptyIfNotExist(s.drive.dock, empty); err != nil {
		return errcode.Annotate(err, "build empty docker image")
	}

	if !config.Bare {
		if err := killOldCoreIfExist(s.drive); err != nil {
			return errcode.Annotate(err, "kill old core")
		}
	}
	d := s.Drive()
	if err := maybeUpdateOS(d); err != nil {
		// Just exit here. If this is a temp error, it will retry the next
		// time the container starts.
		return errcode.Annotate(err, "update os")
	}

	go bg(s)

	const sock = "var/jarvis.sock"
	log.Printf("serve on %s and %s", sock, addr)

	adminService := adminRouter(s)
	go func() {
		if err := aries.ListenAndServe(sock, adminService); err != nil {
			log.Fatal(errcode.Annotate(err, "listen and serve on socket"))
		}
	}()

	service := makeService(s, adminService)
	if err := aries.ListenAndServe(addr, service); err != nil {
		return errcode.Annotate(err, "listen and serve")
	}
	return nil
}

func serverMain() {
	addr := flag.String("addr", "localhost:3377", "address to listen on")
	home := flag.String("home", ".", "home dir")
	flag.Parse()

	if err := run(*home, *addr); err != nil {
		log.Fatal(err)
	}
}
