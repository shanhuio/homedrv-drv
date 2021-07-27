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

package homerelease

import (
	"shanhu.io/aries/creds"
	"shanhu.io/misc/errcode"
)

func cmdPush(server string, args []string) error {
	flags := cmdFlags.New()
	objs := flags.String(
		"objs", "out/homedrv/objs.tar", "path to objects tarball",
	)
	release := flags.String(
		"release", "out/homedrv/release.json", "path to release info",
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

	_ = release

	return nil
}
