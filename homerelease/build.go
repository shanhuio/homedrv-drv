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

package homerelease

import (
	"log"
	"runtime"
	"time"

	"shanhu.io/homedrv/drv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/jsonutil"
	"shanhu.io/misc/jsonx"
	"shanhu.io/misc/semver"
	"shanhu.io/virgo/dock"
)

type builder struct {
	src string
	out string
}

func (b *builder) buildRelease(name, typ string) error {
	switch typ {
	case "":
		typ = "dev"
	case "dev", "prod":
	default:
		return errcode.InvalidArgf("type must be 'dev' or 'prod'")
	}

	arts := new(drvapi.Artifacts)
	const repo = "docker/homedrv"

	log.Println("reading os info")
	osInfoFile := filePath(b.src, repo, "os.jsonx")
	var osInfo struct{ Version string }
	if err := jsonx.ReadFile(osInfoFile, &osInfo); err != nil {
		return errcode.Annotate(err, "read os info")
	}
	arts.OS = osInfo.Version

	log.Println("reading docker pulls")
	pull := make(map[string]string)
	pullFile := filePath(b.src, "docker/homedrv/pull.jsonx")
	if err := jsonx.ReadFile(pullFile, &pull); err != nil {
		return errcode.Annotate(err, "pull tag info")
	}

	images := make(map[string]*dockerImage)
	imageObjs := make(map[string]string)
	for _, d := range []string{
		"nextcloud20",
		"nextcloud21",
		"nextcloud22",
		"postgres12",
		"redis",

		"jarvis",
		"doorway",
		"ncfront",
		"homeboot",
		"toolbox",
	} {
		log.Printf("checksuming %s", d)
		tgz := filePath(b.out, repo, d+".tgz")
		img, err := sumDockerTgz(tgz)
		if err != nil {
			return errcode.Annotatef(err, "checksum for %q", d)
		}
		images[d] = img
		imageObjs[img.sum] = tgz
	}

	log.Println("building artifacts jsonx")
	imageSums := make(map[string]string)
	for _, entry := range []struct {
		name   string
		images []string
		steps  *[]*drvapi.StepVersion
		final  *string
	}{{
		name: "nextcloud",
		images: []string{
			"nextcloud20", "nextcloud21", "nextcloud22",
		},
		steps: &arts.Nextclouds,
		final: &arts.Nextcloud,
	}, {
		name:   "postgres",
		images: []string{"postgres12"},
		steps:  &arts.Postgreses,
		final:  &arts.Postgres,
	}, {
		name:   "redis",
		images: []string{"redis"},
		final:  &arts.Redis,
	}} {
		final := ""
		for _, img := range entry.images {
			id := images[img]
			src := pull[img]
			_, tag := dock.ParseImageTag(src)
			major, err := semver.Major(tag)
			if err != nil {
				return errcode.Annotatef(
					err, "parse tag of %s: %q", img, src,
				)
			}

			if entry.steps != nil {
				step := &drvapi.StepVersion{
					Major:    major,
					Version:  tag,
					Source:   src,
					Image:    id.id,
					ImageSum: id.sum,
				}
				*entry.steps = append(*entry.steps, step)
			}
			final = id.id

			imageSums[id.id] = id.sum
		}
		if entry.final != nil {
			*entry.final = final
		}
	}

	for _, entry := range []struct {
		name string
		id   *string
	}{
		{name: "jarvis", id: &arts.Jarvis},
		{name: "doorway", id: &arts.Doorway},
		{name: "ncfront", id: &arts.NCFront},
		{name: "homeboot", id: &arts.HomeBoot},
		{name: "toolbox", id: &arts.Toolbox},
	} {
		id := images[entry.name]
		*entry.id = id.id
		imageSums[id.id] = id.sum
	}

	arts.ImageSums = imageSums

	log.Printf("writing out artifacts.json")
	artsOut := filePath(b.out, repo, "artifacts.json")
	if err := jsonutil.WriteFileReadable(artsOut, arts); err != nil {
		return errcode.Annotate(err, "write out artifacts")
	}

	log.Printf("writing out release.json")
	rel := &drvapi.Release{
		Name:      name,
		Time:      time.Now(),
		Type:      typ,
		Arch:      runtime.GOARCH,
		Artifacts: arts,
	}
	relOut := filePath(b.out, repo, "/release.json")
	if err := jsonutil.WriteFileReadable(relOut, rel); err != nil {
		return errcode.Annotate(err, "write out release")
	}

	log.Printf("writing out objects")
	objOut := filePath(b.out, repo, "objs.tar")
	if err := writeObjects(objOut, imageObjs); err != nil {
		return errcode.Annotate(err, "writing out object archive")
	}

	return nil
}

func cmdBuild(args []string) error {
	flags := cmdFlags.New()
	src := flags.String("src", "src", "source directory")
	out := flags.String("out", "out", "output directory")
	name := flags.String("name", "", "release name")
	typ := flags.String("type", "", "release type")
	args = flags.ParseArgs(args)

	b := &builder{
		src: *src,
		out: *out,
	}

	return b.buildRelease(*name, *typ)
}
