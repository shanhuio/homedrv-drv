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
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"shanhu.io/misc/errcode"
	"shanhu.io/misc/httputil"
	"shanhu.io/misc/jsonx"
	"shanhu.io/misc/rand"
	"shanhu.io/misc/strutil"
	"shanhu.io/misc/tarutil"
	"shanhu.io/virgo/dock"
)

func randPassword() string {
	return rand.Letters(16)
}

var errSameImage = errors.New("same image")

func dropContIfDifferent(d *dock.Client, name, img string) error {
	c := dock.NewCont(d, name)
	info, err := c.Inspect()
	if err != nil {
		if errcode.IsNotFound(err) {
			log.Printf("container %q not found on upgrading", name)
			return nil
		}
		return errcode.Annotatef(err, "inspect %s", name)
	}
	if info.Image == img {
		return errSameImage // nothing to update
	}
	if err := c.Drop(); err != nil {
		return errcode.Annotatef(err, "drop %s", name)
	}
	return nil
}

func execError(ret int, err error) error {
	if err != nil {
		return err
	}
	if ret != 0 {
		return errcode.Internalf("exit value: %d", ret)
	}
	return nil
}

func addJSONXToTarStream(
	s *tarutil.Stream, f string, m *tarutil.Meta, obj interface{},
) error {
	bs, err := jsonx.Marshal(obj)
	if err != nil {
		return err
	}
	s.AddBytes(f, m, bs)
	return nil
}

func pingDomains(domains []string) {
	set := strutil.MakeSet(domains)
	list := strutil.SortedList(set)

	client := http.DefaultClient

	done := make(map[string]bool)
	for i := 0; i < 3; i++ {
		if len(done) == len(list) {
			break
		}
		for _, d := range list {
			if done[d] {
				continue
			}
			u := &url.URL{Scheme: "https", Host: d}
			code, err := httputil.GetCode(client, u.String())
			if err != nil {
				// TODO(h8liu): investigate why we have EOF error from get.
				if !strings.HasSuffix(err.Error(), ": EOF") {
					log.Printf("warning: ping %s got error: %s", u, err)
				}
			} else if code != http.StatusOK {
				log.Printf("warning: ping %q got status %d", u, code)
			} else {
				done[d] = true
			}
		}
		time.Sleep(time.Second)
	}
}
