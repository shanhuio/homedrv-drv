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
	"io/ioutil"
	"net/url"

	"shanhu.io/misc/errcode"
)

func cmdEnroll(args []string) error {
	flags := cmdFlags.New()
	server := flags.String("server", defaultServer, "server to register")
	name := flags.String("name", "", "endpoint name")
	code := flags.String("code", "", "passcode")
	pubKey := flags.String("pubkey", "", "public key file")
	flags.ParseArgs(args)

	if *name == "" {
		return errcode.InvalidArgf("name not specified")
	}
	if *code == "" {
		return errcode.InvalidArgf("passcode not specified")
	}

	serverURL, err := url.Parse(*server)
	if err != nil {
		return errcode.Annotatef(err, "invalid server url: %q", *server)
	}
	if *pubKey == "" {
		return errcode.InvalidArgf("public key not specified")
	}
	pubKeyBytes, err := ioutil.ReadFile(*pubKey)
	if err != nil {
		return errcode.Annotate(err, "read public key")
	}

	return registerEndpoint(serverURL, *name, *code, pubKeyBytes)
}
