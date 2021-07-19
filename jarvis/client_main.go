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
	"fmt"
	"sort"
	"strings"

	"shanhu.io/homedrv/burmilla"
	"shanhu.io/homedrv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/misc/jsonutil"
	"shanhu.io/misc/nameutil"
	"shanhu.io/misc/subcmd"
	"shanhu.io/pisces/settings"
)

func clientCommands() *subcmd.List {
	c := subcmd.New()
	c.Add("update", "hints to check update", cmdUpdate)
	c.Add("list-os", "list the available os versions", cmdListOS)
	c.Add("occ", "runs occ on nextcloud", cmdOCC)
	c.Add("settings", "prints settings", cmdSettings)
	c.Add("set-password", "sets password of a user", cmdSetPassword)
	c.Add("version", "prints release info", cmdVersion)
	c.Add(
		"nextcloud-domains", "view or modify nextcloud domains",
		cmdNextcloudDomains,
	)
	c.Add(
		"custom-subs", "view or modify additional custom subdomains",
		cmdCustomSubs,
	)
	c.Add("update-os", "upgrade os", cmdUpdateOS)
	c.Add("update-grub-config", "upgrade grub config", cmdUpdateGrubConfig)

	return c
}

func clientMain() { clientCommands().Main() }

func cmdUpdate(args []string) error {
	flags := cmdFlags.New()
	build := flags.String("build", "", "release to install")
	sock := flags.String(
		"sock", "jarvis.sock", "UDS where jarvis is listening",
	)
	args = flags.ParseArgs(args)
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/update", *build, nil)
}

func cmdListOS(args []string) error {
	flags := cmdFlags.New()
	cflags := newClientFlags(flags)
	flags.ParseArgs(args)
	d, err := newClientDrive(cflags)
	if err != nil {
		return errcode.Annotate(err, "init homedrive stub")
	}

	b, err := d.burmilla()
	if err != nil {
		return errcode.Annotate(err, "init burmilla stub")
	}
	lines, err := burmilla.ListOS(b)
	if err != nil {
		return err
	}
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}

func cmdVersion(args []string) error {
	flags := cmdFlags.New()
	cflags := newClientFlags(flags)
	flags.ParseArgs(args)
	d, err := newClientDrive(cflags)
	if err != nil {
		return err
	}

	r := new(drvapi.Release)
	if err := d.settings.Get(keyBuild, r); err != nil {
		return err
	}
	jsonutil.Print(r)
	return nil
}

func cmdSettings(args []string) error {
	flags := cmdFlags.New()
	cflags := newClientFlags(flags)
	flags.ParseArgs(args)
	d, err := newClientDrive(cflags)
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return errcode.Internalf("expects one settings key")
	}

	k := args[0]
	var v interface{}
	if err := d.settings.Get(k, &v); err != nil {
		return err
	}
	jsonutil.Print(v)
	return nil
}

func cmdSetPassword(args []string) error {
	flags := cmdFlags.New()
	sock := flags.String(
		"sock", "jarvis.sock", "jarvis unix domain socket",
	)
	pass := flags.String("pass", "", "password to set")
	args = flags.ParseArgs(args)

	if *pass == "" {
		return errcode.InvalidArgf("new password is empty")
	}
	c := httputil.NewUnixClient(*sock)
	req := &changePasswordRequest{NewPassword: *pass}
	return c.Call("/api/set-password", req, nil)
}

func cmdCustomSubs(args []string) error {
	flags := cmdFlags.New()
	sock := flags.String(
		"sock", "jarvis.sock", "jarvis unix domain socket",
	)
	add := flags.Bool("add", false, "adds a sub domain")
	remove := flags.Bool("remove", false, "removes a sub domain")
	cflags := newClientFlags(flags)
	args = flags.ParseArgs(args)
	list := !*add && !*remove

	d, err := newClientDrive(cflags)
	if err != nil {
		return err
	}

	subMap, err := loadCustomSubs(d.settings)
	if err != nil {
		return errcode.Annotate(err, "read custom subdomains")
	}

	if list {
		if len(args) != 0 {
			return errcode.InvalidArgf("list takes no command")
		}
		subs := []string{}
		for sub := range subMap {
			subs = append(subs, sub)
		}
		sort.Strings(subs)
		for _, sub := range subs {
			fmt.Printf("%s -> %s\n", sub, subMap[sub])
		}
		return nil
	}

	fullDomain := func(sub string) (string, error) {
		// See if user specified a full domain name or just the subdomain.
		idx := strings.Index(sub, ".")
		if idx > 0 {
			// User specified a full domain name. Nothing else to do.
			return sub, nil
		}
		if idx == 0 {
			return "", errcode.InvalidArgf("subdomain can not start with dot")
		}

		// User only specified the custom subdomain label.
		if err := nameutil.CheckLabel(sub); err != nil {
			return "", errcode.Annotate(err, "check subdomain")
		}

		mainDomain, err := settings.String(d.settings, keyMainDomain)
		if err != nil {
			return "", errcode.Annotate(err, "expand subdomain")
		}
		return sub + "." + mainDomain, nil
	}

	if *add {
		if len(args) != 2 {
			return errcode.InvalidArgf("add command takes 2 arguments")
		}
		domain, err := fullDomain(args[0])
		if err != nil {
			return errcode.Annotate(err, "get full domain name")
		}
		dest := args[1]

		if _, ok := subMap[domain]; ok {
			return errcode.InvalidArgf("subdomain %q already exist", domain)
		}
		subMap[domain] = dest
	} else if *remove {
		if len(args) != 1 {
			return errcode.InvalidArgf("remove command takes 1 argument")
		}
		domain, err := fullDomain(args[0])
		if err != nil {
			return errcode.Annotate(err, "get full domain name")
		}
		if _, ok := subMap[domain]; !ok {
			return errcode.InvalidArgf("subdomain %q not in sub list", domain)
		}
		delete(subMap, domain)
	}

	if err := d.settings.Set(keyCustomSubs, subMap); err != nil {
		return errcode.Annotate(err, "save custom subdomain map")
	}

	// Ping jarvis to recreate doorway so that hostmap will be updated.
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/recreate-doorway", nil, nil)
}
