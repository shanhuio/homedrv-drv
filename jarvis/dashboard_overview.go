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
	"net/url"

	"shanhu.io/homedrv/drv/burmilla"
	"shanhu.io/homedrv/drv/homeapp/nextcloud"
	"shanhu.io/pub/errcode"
)

// DashboardOverviewData contains the data for dashboard overview.
type DashboardOverviewData struct {
	NoSysDock       bool
	NextcloudDomain string
	IPAddrs         []string
	DiskUsage       *diskUsage
	UptimeSecs      int64
}

type diskSize struct {
	MB int // x / 1e6
	B  int // x % 1e6
}

func newDiskSize(size uint64) *diskSize {
	return &diskSize{
		MB: int(size / 1e6),
		B:  int(size % 1e6),
	}
}

type diskUsage struct {
	Total *diskSize
	Free  *diskSize
}

func newDashboardOverviewData(s *server) (*DashboardOverviewData, error) {
	d := new(DashboardOverviewData)

	var ncDomains []string
	if err := s.drive.settings.Get(
		nextcloud.KeyDomains, &ncDomains,
	); err != nil {
		if !errcode.IsNotFound(err) {
			return nil, errcode.Internalf("failed to get nextcloud domain")
		}
	}
	if len(ncDomains) > 0 {
		d.NextcloudDomain = (&url.URL{
			Scheme: "https",
			Host:   ncDomains[0],
		}).String()
	}

	if s.drive.sysDock != nil {
		b, err := s.drive.burmilla()
		if err != nil {
			return nil, errcode.Annotate(err, "init burmilla OS stub")
		}

		ips, err := burmilla.HostIPs(b, "eth0")
		if err != nil {
			return nil, errcode.Annotate(err, "get IP address")
		}
		d.IPAddrs = ips

		du, err := burmilla.QueryDiskUsage(b)
		if err != nil {
			return nil, errcode.Annotate(err, "get disk usage")
		}
		d.DiskUsage = &diskUsage{
			Total: newDiskSize(du.Total),
			Free:  newDiskSize(du.Free),
		}

		uptime, err := burmilla.Uptime(b)
		if err != nil {
			return nil, errcode.Annotate(err, "query system uptime")
		}
		d.UptimeSecs = int64(uptime.Seconds())
	} else {
		d.NoSysDock = true
	}
	return d, nil
}
