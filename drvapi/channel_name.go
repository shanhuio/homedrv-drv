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

package drvapi

import (
	"strings"
)

var archs = []string{"amd64", "arm64"}

func archSuffix(arch string) string {
	return "-" + arch
}

// ArchOf returns the architecture of
func ArchOf(name string) string {
	parsed := ParseChannelName(name)
	return parsed.Architecture()
}

// ChannelName is a parsed channel name.
type ChannelName struct {
	Base string
	Arch string
}

// Architecture returns the architecture of the channel.
func (n *ChannelName) Architecture() string {
	if n.Arch == "" {
		return "amd64"
	}
	return n.Arch
}

func (n *ChannelName) String() string {
	if n.Arch == "" {
		return n.Base
	}
	return n.Base + archSuffix(n.Arch)
}

// ParseChannelName parse the channel name into base name and architecture
func ParseChannelName(name string) *ChannelName {
	if name == "" {
		return nil
	}
	for _, arch := range archs {
		suffix := archSuffix(arch)
		if strings.HasSuffix(name, suffix) {
			return &ChannelName{
				Base: strings.TrimSuffix(name, suffix),
				Arch: arch,
			}
		}
	}
	return &ChannelName{Base: name}
}
