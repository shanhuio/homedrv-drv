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

package homeboot

import (
	"log"
	"strings"

	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/virgo/dock"
)

func findNameWithSuffix(suf string, names ...string) (string, bool) {
	for _, name := range names {
		if strings.HasSuffix(name, suf) {
			return name, true
		}
	}
	return "", false
}

func cmdUninstall(args []string) error {
	flags := cmdFlags.New()
	dockerSock := flags.String("docker", "", "docker unix domain socket")
	keepNetwork := flags.Bool("keep_network", false, "keep network")
	keepVolumes := flags.Bool("keep_volumes", false, "keep volumes")
	keepImages := flags.Bool("keep_image", false, "keep images")
	flags.ParseArgs(args)

	d := dock.NewUnixClient(*dockerSock)

	const suffix = drvcfg.DefaultSuffix

	conts, err := dock.ListContsWithLabel(d, drvcfg.LabelName)
	if err != nil {
		return errcode.Annotate(err, "list containers")
	}
	for _, cont := range conts {
		if name, ok := findNameWithSuffix(suffix, cont.Names...); ok {
			log.Printf("remove container %q", name)
			c := dock.NewCont(d, cont.ID)
			if err := c.ForceRemove(); err != nil {
				return errcode.Annotatef(err, "remove container %q", name)
			}
		}
	}

	if !*keepVolumes {
		vols, err := dock.ListVolumesWithLabel(d, drvcfg.LabelName)
		if err != nil {
			return errcode.Annotate(err, "list volumes")
		}
		for _, vol := range vols {
			name := vol.Name
			if !strings.HasSuffix(name, suffix) {
				continue
			}
			log.Printf("remove volume %q", name)
			if err := dock.RemoveVolume(d, name); err != nil {
				return errcode.Annotatef(err, "remove volume %q", name)
			}
		}
	}

	if !*keepNetwork {
		const network = drvcfg.DefaultNetwork
		if has, err := dock.HasNetwork(d, network); err != nil {
			return errcode.Annotate(err, "check network")
		} else if has {
			log.Printf("remove network %q", network)
			if err := dock.RemoveNetwork(d, network); err != nil {
				return errcode.Annotate(err, "remove network")
			}
		}
	}

	if !*keepImages {
		images, err := dock.ListImages(d)
		if err != nil {
			return errcode.Annotate(err, "list images")
		}
		tagPrefix := drvcfg.DefaultRegistry + "/"
		removeOpt := &dock.RemoveImageOptions{}
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if !strings.HasPrefix(tag, tagPrefix) {
					continue
				}
				log.Printf("remove image %q", tag)
				if err := dock.RemoveImage(d, tag, removeOpt); err != nil {
					return errcode.Annotatef(err, "remove image %q", tag)
				}
			}
		}
	}

	log.Println("homedrive uninstalled.")
	return nil
}
