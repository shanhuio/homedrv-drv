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

package homeboot

import (
	"bytes"
	"os"
	"text/template"

	"shanhu.io/g/bosinit"
	"shanhu.io/g/errcode"
	"shanhu.io/g/jsonx"
)

const rcLocal = `#!/bin/bash
exec 1>/var/log/rc.local.log 2>&1

wait-for-docker

set -e

readonly HOMEDRV=/opt/homedrv
{{- $docker := "/var/run/docker.sock" }}
{{- $sysDocker := "/var/run/system-docker.sock" }}

if [[ ! -f "${HOMEDRV}/.init-done" ]]; then
  docker pull {{.HomeBoot}}
  docker run \
    --name "homeboot" \
    --mount "type=bind,source={{$docker}},target={{$docker}}" \
    --mount "type=bind,source={{$sysDocker}},target={{$sysDocker}}" \
    --mount "type=bind,source=${HOMEDRV},target=${HOMEDRV}" \
    {{.HomeBoot}} /bin/homeboot install --config_file="${HOMEDRV}/boot.jsonx"
  date > "${HOMEDRV}/.init-done"
fi
`

var rcLocalTmpl = template.Must(template.New("rc").Parse(rcLocal))

func makeRCLocal(c *InitConfig) (string, error) {
	buf := new(bytes.Buffer)
	if err := rcLocalTmpl.Execute(buf, c); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// InitFiles generates init files for a new homedrive.
func InitFiles(c *InitConfig) ([]*bosinit.WriteFile, error) {
	rc, err := makeRCLocal(c)
	if err != nil {
		return nil, errcode.Annotate(err, "make rc.local file")
	}

	bc, err := jsonx.Marshal(c.Boot)
	if err != nil {
		return nil, errcode.Annotate(err, "make boot.jsonx")
	}

	files := []*bosinit.WriteFile{
		bosinit.BashProfile,
		bosinit.BashRC,
	}
	files = append(files, &bosinit.WriteFile{
		Path:        "/etc/rc.local",
		Permissions: bosinit.FilePerm(0700),
		Owner:       "root",
		Content:     rc,
	}, &bosinit.WriteFile{
		Path:        "/opt/homedrv/boot.jsonx",
		Permissions: bosinit.FilePerm(0644),
		Owner:       "root",
		Content:     string(bc),
	})
	return files, nil
}

func cloudConfig(c *InitConfig) ([]byte, error) {
	files, err := InitFiles(c)
	if err != nil {
		return nil, errcode.Annotate(err, "make files")
	}

	cc := &bosinit.Config{WriteFiles: files}

	keys, err := FetchSSHKeys(c)
	if err != nil {
		return nil, errcode.Annotate(err, "prepare ssh keys")
	}
	cc.SSHAuthorizedKeys = keys

	return cc.CloudConfig()
}

func cmdCloudConfig(args []string) error {
	config := NewInitConfig()
	flags := cmdFlags.New()
	config.DeclareFlags(flags)
	flags.ParseArgs(args)

	if config.Boot.Drive.Name == "" {
		return errcode.InvalidArgf("name not specified")
	}
	if config.Boot.Code == "" {
		return errcode.InvalidArgf("passcode not specified")
	}

	bs, err := cloudConfig(config)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(bs)
	return err
}
