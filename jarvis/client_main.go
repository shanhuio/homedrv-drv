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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"shanhu.io/homedrv/burmilla"
	"shanhu.io/homedrv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/flagutil"
	"shanhu.io/misc/httputil"
	"shanhu.io/misc/jsonutil"
	"shanhu.io/misc/jsonx"
	"shanhu.io/misc/subcmd"
)

func clientCommands() *subcmd.List {
	c := subcmd.New()

	// Jarvis related
	c.Add("version", "prints release info", cmdVersion)
	c.Add("update", "hints to check update", cmdUpdate)
	c.Add("settings", "prints settings", cmdSettings)
	c.Add("set-api-key", "sets API key", cmdSetAPIKey)
	c.Add("set-password", "sets password of a user", cmdSetPassword)
	c.Add("disable-totp", "disables TOTP 2FA", cmdDisableTOTP)
	c.Add(
		"custom-subs", "view or modify additional custom subdomains",
		cmdCustomSubs,
	)

	// OS related
	c.Add("list-os", "list the available os versions", cmdListOS)
	c.Add("update-os", "upgrade os", cmdUpdateOS)
	c.Add("update-grub-config", "upgrade grub config", cmdUpdateGrubConfig)

	// Nextcloud related
	c.Add(
		"set-nextcloud-datamnt", "sets nextcloud data mount point",
		cmdSetNextcloudDataMount,
	)
	c.Add(
		"set-nextcloud-extramnt", "sets nextcloud extra mount points",
		cmdSetNextcloudExtraMount,
	)
	c.Add("nextcloud-cron", "runs nextcloud cron job", cmdNextcloudCron)
	c.Add(
		"nextcloud-domains", "view or modify nextcloud domains",
		cmdNextcloudDomains,
	)

	c.Add("call", "invokes an admin rpc call", cmdCall)

	return c
}

func clientMain() { clientCommands().Main() }

func declareJarvisSockFlag(flags *flagutil.FlagSet) *string {
	return flags.String(
		"sock", "var/jarvis.sock", "jarvis unix domain socket",
	)
}

func cmdUpdate(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	stop := flags.Bool(
		"stop", false, "stop the channel update cron job",
	)
	args = flags.ParseArgs(args)
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/update", !*stop, nil)
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
	sock := declareJarvisSockFlag(flags)
	pass := flags.String("pass", "", "password to set")
	args = flags.ParseArgs(args)

	if *pass == "" {
		return errcode.InvalidArgf("new password is empty")
	}
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/set-password", *pass, nil)
}

func cmdDisableTOTP(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/disable-totp", rootUser, nil)
}

func cmdSetAPIKey(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	keyFile := flags.String("key", "", "key file")
	args = flags.ParseArgs(args)

	if *keyFile == "" {
		return errcode.InvalidArgf("key file is empty")
	}
	key, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		return errcode.Annotate(err, "read key file")
	}
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/set-api-key", key, nil)
}

func cmdSetNextcloudDataMount(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)
	if len(args) != 1 {
		return errcode.InvalidArgf("expect one arg")
	}
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/set-nextcloud-datamnt", args[0], nil)
}

func cmdSetNextcloudExtraMount(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)

	m := make(map[string]string)
	for _, mnt := range args {
		colon := strings.Index(mnt, ":")
		if colon < 0 {
			m[mnt] = mnt
		} else {
			host := mnt[:colon]
			cont := mnt[colon+1:]
			m[host] = cont
		}
	}

	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/set-nextcloud-extramnt", m, nil)
}

func cmdNextcloudCron(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)
	if len(args) != 0 {
		return errcode.InvalidArgf("expect no arg")
	}
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/nextcloud-cron", nil, nil)
}

func cmdCall(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)
	if len(args) == 0 {
		return errcode.InvalidArgf("expect a path to call")
	}
	if len(args) > 2 {
		return errcode.InvalidArgf("too many args")
	}

	c := httputil.NewUnixClient(*sock)

	var req io.Reader
	if len(args) == 1 {
		bs, errs := jsonx.ToJSON([]byte(args[1]))
		if errs != nil {
			return errcode.Annotate(errs[0], "convert request to json")
		}
		req = bytes.NewReader(bs)
	}
	resp := new(bytes.Buffer)
	if err := c.Post(args[0], req, resp); err != nil {
		return err
	}
	respBytes := resp.Bytes()
	bs, err := jsonutil.Format(respBytes)
	if err != nil {
		return errcode.Annotatef(err, "format respose: %s", respBytes)
	}
	fmt.Println(string(bs))
	return nil
}
