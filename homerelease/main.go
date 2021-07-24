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

package homerelease

import (
	"flag"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"shanhu.io/aries/creds"
	"shanhu.io/homedrv/drvapi"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/jsonutil"
	"shanhu.io/misc/jsonx"
	"shanhu.io/misc/rand"
	"shanhu.io/misc/semver"
	"shanhu.io/virgo/dock"
)

func makeReleaseName(typ string) (string, error) {
	ch := typ
	if typ == "dev" {
		u, err := creds.CurrentUser()
		if err != nil {
			return "", errcode.Annotate(err, "get current user")
		}
		ch = "dev-" + u
	}

	date := time.Now().Format("20060102")
	return fmt.Sprintf("%s-%s-%s", ch, date, rand.HexBytes(3)), nil
}

func filePath(base string, parts ...string) string {
	p := path.Join(parts...)
	return filepath.Join(base, filepath.FromSlash(p))
}

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
	if name == "" {
		n, err := makeReleaseName(typ)
		if err != nil {
			return errcode.Annotate(err, "make release name")
		}
		name = n
	}

	arts := new(drvapi.Artifacts)

	log.Println("reading os info")
	osInfoFile := filePath(b.src, "homedrv/os.jsonx")
	var osInfo struct{ Version string }
	if err := jsonx.ReadFile(osInfoFile, &osInfo); err != nil {
		return errcode.Annotate(err, "read os info")
	}
	arts.OS = osInfo.Version

	log.Println("reading docker pulls")
	pull := make(map[string]string)
	pullFile := filePath(b.src, "homedrv/pull.jsonx")
	if err := jsonx.ReadFile(pullFile, &pull); err != nil {
		return errcode.Annotate(err, "pull tag info")
	}

	images := make(map[string]*dockerImage)
	imageObjs := make(map[string]string)
	for _, d := range []string{
		"nextcloud20",
		"nextcloud21",
		"postgres12",
		"redis",
		"influxdb",

		"jarvis",
		"doorway",
		"ncfront",
		"homeboot",
	} {
		log.Printf("checksuming %s", d)
		tgz := filePath(b.out, "homedrv", d+".tgz")
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
		name:   "nextcloud",
		images: []string{"nextcloud20", "nextcloud21"},
		steps:  &arts.Nextclouds,
		final:  &arts.Nextcloud,
	}, {
		name:   "postgres",
		images: []string{"postgres12"},
		steps:  &arts.Postgreses,
		final:  &arts.Postgres,
	}, {
		name:   "redis",
		images: []string{"redis"},
		final:  &arts.Redis,
	}, {
		name:   "influxdb",
		images: []string{"influxdb"},
		final:  &arts.InfluxDB,
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
	} {
		id := images[entry.name]
		*entry.id = id.id
		imageSums[id.id] = id.sum
	}

	arts.ImageSums = imageSums

	log.Printf("writing out artifacts.json")
	artsOut := filePath(b.out, "homedrv/artifacts.json")
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
	relOut := filePath(b.out, "homedrv/release.json")
	if err := jsonutil.WriteFileReadable(relOut, rel); err != nil {
		return errcode.Annotate(err, "write out release")
	}

	log.Printf("writing out objects")
	objOut := filePath(b.out, "homedrv/objs.tar")
	if err := writeObjects(objOut, imageObjs); err != nil {
		return errcode.Annotate(err, "writing out object archive")
	}

	return nil
}

// Main is the main entrance function.
func Main() {
	src := flag.String("src", "src", "source directory")
	out := flag.String("out", "out", "output directory")
	name := flag.String("name", "", "release name; will auto gen when empty")
	typ := flag.String("type", "", "release type")
	flag.Parse()

	b := &builder{
		src: *src,
		out: *out,
	}

	if err := b.buildRelease(*name, *typ); err != nil {
		log.Fatal(err)
	}
}
