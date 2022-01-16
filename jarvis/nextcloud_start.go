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
	"strings"

	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/dock"
)

type nextcloudExtraMount struct {
	host      string
	container string
}

type nextcloudConfig struct {
	domains       []string
	dbPassword    string
	adminPassword string
	redisPassword string
	dataMount     string
	extraMounts   []*nextcloudExtraMount
}

func networkCIDRs(c appCore) ([]string, error) {
	network := appNetwork(c)
	info, err := dock.InspectNetwork(c.Docker(), network)
	if err != nil {
		return nil, err
	}
	if info.IPAM == nil {
		return nil, nil
	}
	var cidrs []string
	for _, c := range info.IPAM.Config {
		cidrs = append(cidrs, c.Subnet)
	}
	return cidrs, nil
}

func nextcloudCreateCont(
	c appCore, image string, config *nextcloudConfig,
) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}
	labels := drvcfg.NewNameLabel(nameNextcloud)
	volName := appVol(c, nameNextcloud)

	contConfig := &dock.ContConfig{
		Name:          appCont(c, nameNextcloud),
		Network:       appNetwork(c),
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Labels:        labels,
	}

	cidrs, err := networkCIDRs(c)
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
	for _, extra := range config.extraMounts {
		contConfig.Mounts = append(contConfig.Mounts, &dock.ContMount{
			Type: dock.MountBind,
			Host: extra.host,
			Cont: extra.container,
		})
	}
	contConfig.Env = map[string]string{
		"POSTGRES_HOST":       appCont(c, namePostgres),
		"POSTGRES_DB":         "nextcloud",
		"POSTGRES_USER":       "nextcloud",
		"POSTGRES_PASSWORD":   config.dbPassword,
		"REDIS_HOST":          appCont(c, nameRedis),
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

	d := c.Docker()
	if _, err := dock.CreateVolumeIfNotExist(
		d, volName, &dock.VolumeConfig{Labels: labels},
	); err != nil {
		return nil, errcode.Annotate(err, "create volume")
	}
	return dock.CreateCont(d, image, contConfig)
}

func nextcloudStart(
	c appCore, image string, config *nextcloudConfig,
) error {
	cont, err := nextcloudCreateCont(c, image, config)
	if err != nil {
		return errcode.Annotate(err, "create nextcloud")
	}
	return cont.Start()
}
