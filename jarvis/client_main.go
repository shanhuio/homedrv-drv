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
	"strings"

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

	c.Add("call", "invokes an admin rpc call", cmdCall)

	// Jarvis related
	c.Add("version", "prints release info", cmdVersion)
	c.Add("settings", "prints settings", cmdSettings)

	c.Add("update", "hints to check update", cmdUpdate)
	c.Add("set-password", "sets password of a user", cmdSetPassword)
	c.Add("disable-totp", "disables TOTP 2FA", cmdDisableTOTP)
	c.Add(
		"custom-subs", "view or modify additional custom subdomains",
		cmdCustomSubs,
	)

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
	stop := flags.Bool("stop", false, "stop the channel update cron job")
	args = flags.ParseArgs(args)
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/admin/update", !*stop, nil)
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
	return c.Call("/api/admin/set-password", *pass, nil)
}

func cmdDisableTOTP(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/admin/disable-totp", rootUser, nil)
}

func cmdSetNextcloudDataMount(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)
	if len(args) != 1 {
		return errcode.InvalidArgf("expect one arg")
	}
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/admin/set-nextcloud-datamnt", args[0], nil)
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
	return c.Call("/api/admin/set-nextcloud-extramnt", m, nil)
}

func cmdNextcloudCron(args []string) error {
	flags := cmdFlags.New()
	sock := declareJarvisSockFlag(flags)
	args = flags.ParseArgs(args)
	if len(args) != 0 {
		return errcode.InvalidArgf("expect no arg")
	}
	c := httputil.NewUnixClient(*sock)
	return c.Call("/api/admin/nextcloud-cron", nil, nil)
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
