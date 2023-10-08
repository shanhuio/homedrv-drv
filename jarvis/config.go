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
	"shanhu.io/g/jsonx"
	"shanhu.io/g/osutil"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
)

func readConfig(h *osutil.Home) (*drvcfg.Config, error) {
	f := h.Var("config.jsonx")
	c := new(drvcfg.Config)
	if err := jsonx.ReadFile(f, c); err != nil {
		return nil, err
	}
	return c, nil
}
