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

package drvconfig

import (
	"path"
)

// Naming defines the naming conventions of a jarvis installation.
type Naming struct {
	// Network name. Default: "homedrv"
	Network string `json:",omitempty"`

	// Suffix for container and volume names. Default: ".homedrv"
	Suffix string `json:",omitempty"`

	// Image registry path of downloaded images. Default: "cr.homedrive.io"
	Registry string `json:",omitempty"`
}

// Default names
const (
	DefaultNetwork  = "homedrv"
	DefaultSuffix   = ".homedrv"
	DefaultRegistry = "cr.homedrive.io"
)

// Name returns the name of a container or volume.
func Name(n *Naming, cont string) string {
	if n == nil {
		return cont
	}
	suffix := n.Suffix
	if suffix == "" {
		suffix = DefaultSuffix
	}
	return cont + suffix
}

// Image returns the image name of an image type.
func Image(n *Naming, img string) string {
	if n == nil {
		switch img {
		case "jarvis", "doorway":
			// TODO(h8liu): deprecate this tagging
			return path.Join("registry.digitalocean.com/shanhu", img)
		default:
			return img
		}
	}

	reg := n.Registry
	if reg == "" {
		reg = DefaultRegistry
	}
	const project = "homedrv"
	return path.Join(reg, project, img)
}

// Core returns the name of the core.
func Core(n *Naming) string {
	if n == nil {
		return "jarvis"
	}
	return Name(n, "core")
}

// OldCore returns the name of the old core.
func OldCore(n *Naming) string {
	if n == nil {
		return "jarvis-old"
	}
	return Name(n, "old.core")
}

// Network returns the network's name.
func Network(n *Naming) string {
	if n == nil || n.Network == "" {
		return DefaultNetwork
	}
	return n.Network
}
