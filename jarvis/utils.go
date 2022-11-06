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
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"shanhu.io/pub/httputil"
	"shanhu.io/pub/jsonx"
	"shanhu.io/pub/strutil"
	"shanhu.io/pub/tarutil"
)

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
