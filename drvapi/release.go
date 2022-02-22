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
	"time"
)

// Release is a set of release artifacts with meta data like name, type and
// time.
type Release struct {
	Name string
	Type string
	Time time.Time
	Arch string `json:",omitempty"`

	*Artifacts

	Apps []*AppMeta `json:",omitempty"`
}

// EmptyRelease returns an empty release.
func EmptyRelease() *Release {
	return &Release{Artifacts: &Artifacts{}}
}

// Artifacts contains a set of artifacts and docker images for a release.
type Artifacts struct {
	OS string // OS version.

	// Deprecated. Rancher OS release version.
	RancherOS string `json:",omitempty"`

	// Docker images
	Jarvis  string
	Doorway string
	Toolbox string

	// HomeBoot is saved for provisioning only.
	HomeBoot string

	// App images.
	NCFront string

	// 3rd-party app images.
	Nextcloud string
	Redis     string
	Postgres  string

	// Upgrade path for nextcloud
	Nextclouds []*StepVersion `json:",omitempty"`
	Postgreses []*StepVersion `json:",omitempty"`

	// Checksums for images
	ImageSums map[string]string `json:",omitempty"`
}

// UpdateQueryRequest is a query for asking for latest update.
type UpdateQueryRequest struct {
	Channel      string
	CurrentBuild string `json:",omitempty"`
	Tags         string `json:",omitempty"`

	Manual bool `json:",omitempty"`
}

// UpdateQueryResponse is the response for an update query.
type UpdateQueryResponse struct {
	Release       *Release `json:",omitempty"`
	AlreadyLatest bool     `json:",omitempty"`
}

func lastStepVersion(steps []*StepVersion) string {
	if len(steps) == 0 {
		return ""
	}
	last := steps[len(steps)-1]
	return last.Version
}

// LegacyAppsFromArtifacts returns the apps that that is implied by the
// release. These are for releases that does not have the Apps filled.
func LegacyAppsFromArtifacts(arts *Artifacts) []*AppMeta {
	var metas []*AppMeta

	const (
		appRedis          = "redis"
		appPostgres       = "postgres"
		appNextcloudFront = "ncfront"
		appNextcloud      = "nextcloud"
	)

	for _, m := range []*AppMeta{{
		Name:  appRedis,
		Image: arts.Redis,
	}, {
		Name:  appPostgres,
		Image: arts.Postgres,
		Steps: arts.Postgreses,
	}, {
		Name:  appNextcloudFront,
		Image: arts.NCFront,
	}, {
		Name: appNextcloud,
		Deps: []string{
			appNextcloudFront,
			appPostgres,
			appRedis,
		},
		Image:      arts.Nextcloud,
		SemVersion: lastStepVersion(arts.Nextclouds),
		Steps:      arts.Nextclouds,
	}} {
		if m.Image != "" {
			metas = append(metas, m)
		}
	}

	return metas
}
