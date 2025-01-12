// Copyright (C) 2023  Shanhu Tech Inc.
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
	"shanhu.io/g/flagutil"
)

// InitConfig is the configuration to initialize a homedrive.
type InitConfig struct {
	HomeBoot string

	Boot *BootConfig

	GitHubKeys string
	UserKeys   string
}

// NewInitConfig creates a new init config for creating a new
// HomeDrive instance.
func NewInitConfig() *InitConfig {
	return &InitConfig{Boot: newBootConfig()}
}

// DeclareFlags declares command line flags.
func (c *InitConfig) DeclareFlags(flags *flagutil.FlagSet) {
	c.Boot.declareFlags(flags)

	flags.StringVar(
		&c.HomeBoot, "homeboot", "homedrv/homeboot",
		"init docker image",
	)
	flags.StringVar(
		&c.GitHubKeys, "github_keys", "",
		"add the ssh keys of a github user",
	)
	flags.StringVar(
		&c.UserKeys, "user_keys", "",
		"add the ssh keys of a homedrive user",
	)
}
