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

package homeboot

import (
	"log"

	"shanhu.io/homedrv/drvapi"
	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/semver"
)

// DownloadConfig is the install config. This is the configuration
// for downloading and installing.
type DownloadConfig struct {
	Release *drvapi.Release
	Channel string
	Build   string

	Naming *drvcfg.Naming // Naming conventions.

	// Download the core only; only used in homeboot for bootstraping.
	CoreOnly   bool
	LatestOnly bool

	// If set, ignore major versions that are lower than this.
	CurrentSemVersions map[string]string
}

func (c *DownloadConfig) currentMajor(app string) int {
	if c.CurrentSemVersions == nil {
		return 0
	}
	v, ok := c.CurrentSemVersions[app]
	if !ok {
		return 0
	}
	if v == "" {
		return 0
	}
	major, err := semver.Major(v)
	if err != nil {
		log.Printf("invalid sem version of %q: %q: %s", app, v, err)
		return 0
	}
	return major
}
