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

package homerelease

import (
	"encoding/json"

	"shanhu.io/aries/creds"
	"shanhu.io/homedrv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/jsonutil"
)

func cmdPush(server string, args []string) error {
	flags := cmdFlags.New()
	objs := flags.String(
		"objs", "out/docker/homedrv/objs.tar", "path to objects tarball",
	)
	rel := flags.String(
		"release", "out/docker/homedrv/release.json", "path to release info",
	)
	user := flags.String(
		"user", "root", "user to call the push API",
	)
	args = flags.ParseArgs(args)

	c, err := creds.DialAsUser(*user, server)
	if err != nil {
		return errcode.Annotate(err, "dial server")
	}

	up := &Uploader{
		Client:  c,
		DataURL: "/obj",
		APIURL:  "/api/obj",
	}
	if err := up.Upload(*objs); err != nil {
		return errcode.Annotate(err, "upload objects")
	}

	release := new(drvapi.Release)
	if err := jsonutil.ReadFile(*rel, release); err != nil {
		return errcode.Annotate(err, "read release file")
	}
	newName, err := MakeReleaseName(release.Type)
	if err != nil {
		return errcode.Annotate(err, "make release name")
	}
	release.Name = newName

	bs, err := json.Marshal(release)
	if err != nil {
		return errcode.Annotate(err, "marshal  release")
	}
	return c.Call("/api/sys/push-update", bs, nil)
}
