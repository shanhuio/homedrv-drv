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
	"log"

	"shanhu.io/misc/errcode"
)

func fixThings(d *drive) {
	if d.config.Bare {
		return
	}

	if err := fixOSUpgradeURL(d); err != nil {
		log.Println("fix os upgrade url: ", err)
	}
	if err := fixNextcloud(d); err != nil {
		log.Println("fix nextcloud: ", err)
	}
}

func fixNextcloud(d *drive) error {
	nc := newNextcloud(d)
	return nc.fix()
}

func fixOSUpgradeURL(d *drive) error {
	if !isOSUpdateSupported(d) {
		return nil
	}
	b, err := d.burmilla()
	if err != nil {
		return errcode.Annotate(err, "init os stub")
	}
	return setOSUpdateSource(b)
}
