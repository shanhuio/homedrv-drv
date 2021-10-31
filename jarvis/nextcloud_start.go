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
	"strings"

	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/dock"
)

type nextcloudConfig struct {
	domains       []string
	dbPassword    string
	adminPassword string
	redisPassword string
	dataMount     string
}

func nextcloudCreateCont(
	drive *drive, d *dock.Client, image string, config *nextcloudConfig,
) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}
	labels := drvcfg.NewNameLabel(nameNextcloud)
	volName := drive.vol(nameNextcloud)

	contConfig := &dock.ContConfig{
		Name:          drive.cont(nameNextcloud),
		Network:       drive.network(),
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Labels:        labels,
	}

	cidrs, err := networkCIDRs(drive)
	if err != nil {
		return nil, errcode.Annotate(err, "list network CIDRs")
	}

	contConfig.Mounts = append(contConfig.Mounts, &dock.ContMount{
		Type: dock.MountVolume,
		Host: volName,
		Cont: "/var/www/html",
	})
	if config.dataMount != "" {
		contConfig.Mounts = append(contConfig.Mounts, &dock.ContMount{
			Type: dock.MountBind,
			Host: config.dataMount,
			Cont: "/var/www/html/data",
		})
	}
	contConfig.Env = map[string]string{
		"POSTGRES_HOST":       drive.cont(namePostgres),
		"POSTGRES_DB":         "nextcloud",
		"POSTGRES_USER":       "nextcloud",
		"POSTGRES_PASSWORD":   config.dbPassword,
		"REDIS_HOST":          drive.cont(nameRedis),
		"REDIS_HOST_PASSWORD": config.redisPassword,

		"NEXTCLOUD_ADMIN_USER":     "admin",
		"NEXTCLOUD_ADMIN_PASSWORD": config.adminPassword,
	}
	if len(config.domains) > 0 {
		domains := strings.Join(config.domains, " ")
		contConfig.Env["NEXTCLOUD_TRUSTED_DOMAINS"] = domains
	}
	if len(cidrs) > 0 {
		proxies := strings.Join(cidrs, " ")
		contConfig.Env["TRUSTED_PROXIES"] = proxies
	}

	if _, err := dock.CreateVolumeIfNotExist(
		d, volName, &dock.VolumeConfig{Labels: labels},
	); err != nil {
		return nil, errcode.Annotate(err, "create volume")
	}
	return dock.CreateCont(d, image, contConfig)
}

func nextcloudStart(
	drive *drive, d *dock.Client, image string,
	config *nextcloudConfig,
) error {
	cont, err := nextcloudCreateCont(drive, d, image, config)
	if err != nil {
		return errcode.Annotate(err, "create nextcloud")
	}
	return cont.Start()
}
