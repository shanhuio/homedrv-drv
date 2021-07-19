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
	"testing"

	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"

	"shanhu.io/aries/https/httpstest"
	"shanhu.io/misc/jsonx"
)

func checkGet(t *testing.T, c *http.Client, url, want string) {
	resp, err := c.Get(url)
	if err != nil {
		t.Errorf("get %s: %s", url, err)
		return
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read body: %s", err)
		return
	}

	if string(bs) != want {
		t.Errorf("get %s, want %q, got %q", url, want, string(bs))
	}
}

func TestServe(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf("request: %q", req.URL)
		for k := range req.Header {
			v := req.Header.Get(k)
			t.Logf("%q=%q", k, v)
		}
		fmt.Fprint(w, "dest")
	})
	s := httptest.NewServer(h)

	hostMap := map[string]string{
		"ctrl.shanhu.io": HomeHost,
		"shanhu.io":      lisAddr(s.Listener),
	}

	doorwayHome, err := ioutil.TempDir("", "doorway")
	if err != nil {
		t.Fatal("make doorway temp home:", err)
	}
	defer os.RemoveAll(doorwayHome)

	doorwayEtc := filepath.Join(doorwayHome, "etc")
	if err := os.MkdirAll(doorwayEtc, 0700); err != nil {
		t.Fatal("make doorway etc:", err)
	}
	doorwayVar := filepath.Join(doorwayHome, "var")
	if err := os.MkdirAll(doorwayVar, 0700); err != nil {
		t.Fatal("make doorway var:", err)
	}

	mapFile := filepath.Join(doorwayEtc, "host-map.jsonx")
	if err := jsonx.WriteFile(mapFile, hostMap); err != nil {
		t.Fatal("write host map", err)
	}

	tlsConfigs, err := httpstest.NewTLSConfigs(
		[]string{"ctrl.shanhu.io", "shanhu.io"},
	)
	if err != nil {
		t.Fatal(err)
	}
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	config, err := ConfigFromDirs(doorwayEtc, doorwayVar)
	if err != nil {
		t.Fatal("read config:", err)
	}
	internal := makeInternalConfig(config)
	internal.listen = &listenConfig{
		local: &localListenConfig{listener: lis},
	}
	internal.tlsConfig = tlsConfigs.Server

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bgErr := make(chan error, 1)
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		bgErr <- serve(ctx, internal)
	}(ctx)

	client := &http.Client{
		Transport: tlsConfigs.Sink(lisAddr(lis)),
	}
	checkGet(t, client, "https://shanhu.io", "dest")
	checkGet(t, client, "https://shanhu.io/subpage", "dest")
	checkGet(t, client, "https://ctrl.shanhu.io/health", "ok")

	cancel()
	if err := <-bgErr; err != nil {
		if err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}
}
