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
	"fmt"

	"shanhu.io/misc/errcode"
	"shanhu.io/pisces/settings"
)

func setNextcloudDomainsIfNotExist(d *drive, domains []string) error {
	ok, err := d.settings.Has(keyNextcloudDomains)
	if err != nil {
		return errcode.Annotate(err, "check nextcloud domain")
	}
	if ok {
		return nil
	}
	return d.settings.Set(keyNextcloudDomains, domains)
}

func nextcloudDomains(d *drive) ([]string, error) {
	var domains []string
	if err := d.settings.Get(keyNextcloudDomains, &domains); err == nil {
		return domains, nil
	} else if !errcode.IsNotFound(err) {
		return nil, err
	}
	// Domain list not found.

	set := func(domains []string) ([]string, error) {
		if err := d.settings.Set(keyNextcloudDomains, domains); err != nil {
			return nil, errcode.Annotate(err, "set nextcloud domains")
		}
		return domains, nil
	}

	domain, err := settings.String(d.settings, keyNextcloudDomain)
	if err == nil {
		return set([]string{domain})
	}
	if !errcode.IsNotFound(err) {
		return nil, errcode.Annotate(err, "read nextcloud domain")
	}
	// Single domain not found.

	main, err := settings.String(d.settings, keyMainDomain)
	if err != nil {
		return nil, errcode.Annotate(err, "cannot determine domain")
	}
	return set([]string{fmt.Sprintf("nextcloud.%s", main)})
}

func loadNextcloudConfig(d *drive) (*nextcloudConfig, error) {
	// TODO(h8liu): reading redis password should to go redis?
	redisPass, err := settings.String(d.settings, keyRedisPass)
	if err != nil {
		return nil, errcode.Annotate(err, "read redis password")
	}

	adminPass, err := readPasswordOrSetRandom(
		d.settings, keyNextcloudAdminPass,
	)
	if err != nil {
		return nil, errcode.Annotate(err, "read init password")
	}
	dbPass, err := readPasswordOrSetRandom(d.settings, keyNextcloudDBPass)
	if err != nil {
		return nil, errcode.Annotate(err, "read db password")
	}
	domains, err := nextcloudDomains(d)
	if err != nil {
		return nil, errcode.Annotate(err, "load domains")
	}
	return &nextcloudConfig{
		domains:       domains,
		dbPassword:    dbPass,
		adminPassword: adminPass,
		redisPassword: redisPass,
	}, nil
}
