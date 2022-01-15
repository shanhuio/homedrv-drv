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
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	drvcfg "shanhu.io/homedrv/drvconfig"
	"shanhu.io/misc/errcode"
	"shanhu.io/misc/osutil"
	"shanhu.io/virgo/dock"
)

var grubConfigTmpl = template.Must(template.New("grubcfg").Parse(`
if loadfont /boot/grub/font.pf2 ; then
	set gfxmode=auto
	insmod efi_gop
	insmod efi_uga
	insmod gfxterm
	terminal_output gfxterm
fi

set menu_color_normal=white/black
set menu_color_highlight=black/light-gray

set timeout=5
menuentry "HomeDrive OS" {
	search --no-floppy --set=root --label RANCHER_STATE
	linux	{{.Vmlinuz}}  {{.BootArgs}}
	initrd	{{.Initrd}}
}
`))

var defaultBootArgs = strings.Join([]string{
	"printk.devkmsg=on",
	"rancher.state.dev=LABEL=RANCHER_STATE",
	"rancher.state.wait",
	"panic=10",
	"console=tty0",
	"rancher.autologin=tty1",
}, " ")

type grubConfig struct {
	Vmlinuz  string
	BootArgs string
	Initrd   string
}

var grubConfigMap = map[string]*grubConfig{
	"v1.5.6": {
		Vmlinuz:  "/boot/vmlinuz-4.14.138-rancher",
		BootArgs: defaultBootArgs,
		Initrd:   "/boot/initrd-v1.5.6",
	},
	"burmilla/os:v1.9.1": {
		Vmlinuz:  "/boot/vmlinuz-4.14.218-burmilla",
		BootArgs: defaultBootArgs,
		Initrd:   "/boot/initrd-v1.9.1",
	},
}

func makeGrubConfig(osVersion string) ([]byte, error) {
	c, found := grubConfigMap[osVersion]
	if !found {
		return nil, errcode.InvalidArgf("unrecognized os: %q", osVersion)
	}

	buf := new(bytes.Buffer)
	if err := grubConfigTmpl.Execute(buf, c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func updateBootPart(dev, osVersion string) error {
	configContent, err := makeGrubConfig(osVersion)
	if err != nil {
		return errcode.Annotate(err, "make grub config")
	}

	const mnt = "/mnt/bootpart"
	mntExist, err := osutil.IsDir(mnt)
	if err != nil {
		return errcode.Annotatef(err, "check if %q exists", mnt)
	}
	if !mntExist {
		if err := os.Mkdir(mnt, 0700); err != nil {
			return errcode.Annotate(err, "make mount dir")
		}
	}
	defer os.RemoveAll(mnt)

	if err := mountBootPart(dev, mnt); err != nil {
		return errcode.Annotate(err, "mount boot partition")
	}
	defer func() {
		if err := unmountBootPart(mnt); err != nil {
			log.Printf("unmount %q: %s", mnt, err)
		}
	}()

	grubConfig := filepath.Join(mnt, "boot/grub/grub.cfg")
	found, err := osutil.IsRegular(grubConfig)
	if err != nil {
		return errcode.Annotate(err, "check grub config file")
	}
	if !found {
		return errcode.Internalf("%q not found", grubConfig)
	}

	swap := filepath.Join(mnt, "boot/grub/grub.cfg.swap")
	if err := ioutil.WriteFile(swap, configContent, 0755); err != nil {
		return errcode.Annotate(err, "write grub config")
	}
	if err := os.Rename(swap, grubConfig); err != nil {
		return errcode.Annotate(err, "apply grub config")
	}

	return nil
}

func runUpdateBootPart(d *drive, osVersion string) error {
	self, err := dock.InspectCont(d.dock, d.core())
	if err != nil {
		return errcode.Annotate(err, "inspect self")
	}
	img := self.Image

	// Run current jarvis docker image in priviledged mode to
	// update grub config files. Send in the dev and osVersion as args.
	const dev = "/dev/sda1" // boot partition device on nuc7

	dockConfig := &dock.ContConfig{
		Privileged: true,
		Devices:    []*dock.ContDevice{{Host: dev}},
		Labels:     drvcfg.NewNameLabel("temp"),
		Cmd: []string{
			"jarvis", "update-grub-config",
			"-dev", dev,
			"-os", osVersion,
		},
	}

	cont, err := dock.CreateCont(d.dock, img, dockConfig)
	if err != nil {
		return err
	}
	defer cont.Drop()

	if err := cont.Start(); err != nil {
		return errcode.Annotate(err, "start boot partition update")
	}

	ret, err := cont.Wait(dock.NotRunning)
	if err != nil {
		return errcode.Annotate(err, "waiting for update to finish")
	}
	if ret != 0 {
		return errcode.Internalf("boot partition update returned: %d", ret)
	}

	return nil
}
