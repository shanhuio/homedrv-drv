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

package burmilla

import (
	"bytes"
	"text/template"

	"shanhu.io/misc/errcode"
)

const mkdirCmdTmpl = `
[[ -d {{.Dir}} ]] || 
(mkdir -m 0700 {{.Dir}} && chown {{.User}}:{{.User}} {{.Dir}})
`

func mkdirCmd(dir, user string) (string, error) {
	// No-op if the directory already exists.
	t := template.Must(template.New("mkdir").Parse(mkdirCmdTmpl))
	d := struct {
		Dir  string
		User string
	}{
		Dir:  dir,
		User: user,
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, d); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Mkdir creates a directory in the console's file system.
func Mkdir(b *Burmilla, dir, user string) error {
	// Make sure /home/rancher/.ssh exists.
	mkdir, err := mkdirCmd(dir, user)
	if err != nil {
		return errcode.Annotate(err, "build mkdir script")
	}
	return execError(b.ExecRet([]string{
		"/bin/bash", "-c", mkdir,
	}))
}
