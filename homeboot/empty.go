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
	"shanhu.io/pub/dock"
)

const emptyDockerFile = `FROM scratch
MAINTAINER Shanhu Tech Inc.
CMD ["/bin/sleep", "1"]
`

// BuildEmpty builds the homedrv/empty image. This image is only used for
// processing volumes.
func BuildEmpty(client *dock.Client, name string) error {
	files := dock.NewTarStream(emptyDockerFile)
	return dock.BuildImageStream(client, name, files)
}

// BuildEmptyIfNotExist builds the homedrv/empty image if the image
// does not exist yet.
func BuildEmptyIfNotExist(client *dock.Client, name string) error {
	has, err := dock.HasImage(client, name)
	if err != nil {
		return err
	}
	if !has {
		return BuildEmpty(client, name)
	}
	return nil
}
