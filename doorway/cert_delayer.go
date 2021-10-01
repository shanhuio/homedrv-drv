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

package doorway

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	"sync"
	"time"
)

type certTimeEntry struct {
	firstSeen time.Time
	expire    time.Time
}

type getCertFunc func(hello *tls.ClientHelloInfo) (*tls.Certificate, error)

type certGetter struct {
	getFunc getCertFunc

	mu          sync.Mutex
	certs       map[string]*certTimeEntry
	manual      map[string]*tls.Certificate
	nextCleanUp time.Time
}

func newCertGetter(
	f getCertFunc, manual map[string]*tls.Certificate,
) *certGetter {
	return &certGetter{
		getFunc:     f,
		certs:       make(map[string]*certTimeEntry),
		nextCleanUp: time.Now().Add(time.Hour),
	}
}

func (g *certGetter) delay(cert *x509.Certificate) {
	now := time.Now()
	if cert.NotBefore.Before(now.Add(-2 * time.Hour)) {
		// cert valid start time is more than 2 hours ago.
		// this is not likely a new certificate.
		return
	}

	k := fmt.Sprintf("%x", cert.SerialNumber)

	g.mu.Lock()
	defer g.mu.Unlock()

	const delay = 2 * time.Second
	entry, ok := g.certs[k]
	if !ok {
		time.Sleep(delay)
		g.certs[k] = &certTimeEntry{
			firstSeen: now,
			expire:    cert.NotAfter,
		}
	} else if now.Before(entry.firstSeen.Add(3 * time.Second)) {
		time.Sleep(delay)
	}

	if now.After(g.nextCleanUp) {
		var toDelete []string
		for k, v := range g.certs {
			if now.After(v.expire) {
				toDelete = append(toDelete, k)
			}
		}
		for _, k := range toDelete {
			delete(g.certs, k)
		}
		g.nextCleanUp = now.Add(time.Hour)
	}
}

func (g *certGetter) get(hello *tls.ClientHelloInfo) (
	*tls.Certificate, error,
) {
	if g.manual != nil {
		name := strings.TrimSuffix(hello.ServerName, ".")
		if cert, ok := g.manual[name]; ok {
			return cert, nil
		}
	}

	cert, err := g.getFunc(hello)
	if err != nil {
		return cert, err
	}
	if cert.Leaf != nil {
		g.delay(cert.Leaf)
	}
	return cert, nil
}
