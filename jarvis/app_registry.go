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
	"shanhu.io/homedrv/drvapi"
	"shanhu.io/homedrv/homeapp/nextcloud"
	"shanhu.io/homedrv/homeapp/postgres"
	"shanhu.io/homedrv/homeapp/redis"
	"shanhu.io/misc/errcode"
)

type appRegistry struct {
	manifest map[string]*drvapi.AppMeta
}

func lastStepVersion(steps []*drvapi.StepVersion) string {
	if len(steps) == 0 {
		return ""
	}
	last := steps[len(steps)-1]
	return last.Version
}

func manifestFromRelease(rel *drvapi.Release) map[string]*drvapi.AppMeta {
	if rel == nil {
		return make(map[string]*drvapi.AppMeta)
	}
	metas := rel.Apps
	if metas == nil {
		if rel.Artifacts == nil {
			return make(map[string]*drvapi.AppMeta)
		}
		for _, m := range []*drvapi.AppMeta{{
			Name:  redis.Name,
			Image: rel.Redis,
		}, {
			Name:  postgres.Name,
			Image: rel.Postgres,
			Steps: rel.Postgreses,
		}, {
			Name:  nextcloud.NameFront,
			Image: rel.NCFront,
		}, {
			Name: nextcloud.Name,
			Deps: []string{
				nextcloud.NameFront,
				postgres.Name,
				redis.Name,
			},
			Image:      rel.Nextcloud,
			SemVersion: lastStepVersion(rel.Nextclouds),
			Steps:      rel.Nextclouds,
		}} {
			if m.Image != "" {
				metas = append(metas, m)
			}
		}
	}
	return makeManifest(metas)
}

func newAppRegistry(rel *drvapi.Release) *appRegistry {
	manifest := manifestFromRelease(rel)
	return &appRegistry{
		manifest: manifest,
	}
}

func (r *appRegistry) meta(name string) (*drvapi.AppMeta, error) {
	meta, found := r.manifest[name]
	if !found {
		return nil, errcode.NotFoundf("app meta not found for %q", name)
	}
	return meta, nil
}

func (r *appRegistry) setRelease(rel *drvapi.Release) {
	r.manifest = manifestFromRelease(rel)
}
