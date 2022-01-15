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
	"strings"

	"shanhu.io/aries"
	"shanhu.io/homedrv/burmilla"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/tarutil"
	"shanhu.io/virgo/bosinit"
)

type sshKeys struct {
	drive *drive
}

func newSSHKeys(d *drive) *sshKeys {
	return &sshKeys{drive: d}
}

type updateSSHKeysRequest struct {
	Keys string
}

type updateSSHKeysResponse struct{}

func (k *sshKeys) list() ([]string, error) {
	b, err := k.drive.burmilla()
	if err != nil {
		return nil, errcode.Annotate(err, "create os stub")
	}
	config, err := burmilla.ConfigExport(b)
	if err != nil {
		return nil, errcode.Annotate(err, "get ssh keys")
	}
	return config.SSHAuthorizedKeys, nil
}

func (k *sshKeys) apiUpdate(c *aries.C, req *updateSSHKeysRequest) (
	*updateSSHKeysResponse, error,
) {
	b, err := k.drive.burmilla()
	if err != nil {
		return nil, errcode.Annotate(err, "create os stub")
	}

	var keys []string
	for _, line := range strings.Split(req.Keys, "\n") {
		if k := strings.TrimSpace(line); k != "" {
			// TODO(jungong) : do we need to make sure all the non-empty
			// keys here are parsable?
			keys = append(keys, k)
		}
	}

	config := &bosinit.Config{SSHAuthorizedKeys: keys}
	if err := burmilla.ConfigMerge(b, config); err != nil {
		return nil, errcode.Annotate(err, "merge config")
	}

	// Update authorized_keys file.
	const (
		uname   = "rancher"
		dir     = "/home/rancher/.ssh"
		keyFile = "authorized_keys"
	)

	uid, err := burmilla.UserID(b, uname)
	if err != nil {
		return nil, errcode.Annotate(err, "get uid")
	}
	gid, err := burmilla.GroupID(b, uname)
	if err != nil {
		return nil, errcode.Annotate(err, "get gid")
	}
	if err := burmilla.Mkdir(b, dir, uname); err != nil {
		return nil, errcode.Annotate(err, "make .ssh directory")
	}

	stream := tarutil.NewStream()
	meta := &tarutil.Meta{
		Mode:    0600,
		UserID:  uid,
		GroupID: gid,
	}
	stream.AddString(keyFile, meta, strings.Join(keys, "\n")+"\n")
	if err := b.CopyInTarStream(stream, dir); err != nil {
		return nil, errcode.Annotate(err, "write authorized_keys")
	}

	return &updateSSHKeysResponse{}, nil
}

func (k *sshKeys) api() *aries.Router {
	r := aries.NewRouter()
	r.Call("update", k.apiUpdate)
	return r
}
