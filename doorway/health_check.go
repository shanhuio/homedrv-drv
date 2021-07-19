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
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type healthChecker struct {
	client  *http.Client
	url     *url.URL
	timeout time.Duration
}

func (c *healthChecker) check(ctx C) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &http.Request{
		Method: "GET",
		URL:    c.url,
	}
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d - %s", resp.StatusCode, resp.Status)
	}
	return nil
}

func healthCheck(ctx C, host string) error {
	const interval = 10 * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fails := 0

	checker := &healthChecker{
		client: new(http.Client),
		url: &url.URL{
			Scheme: "https",
			Host:   host,
		},
		timeout: interval,
	}

	for range ticker.C {
		if err := checker.check(ctx); err != nil {
			log.Printf("health check: %s", err)

			fails++
			if fails >= 5 {
				return fmt.Errorf("failed for 5 times")
			}
			continue
		}

		fails = 0
	}

	panic("unreachable")
}
