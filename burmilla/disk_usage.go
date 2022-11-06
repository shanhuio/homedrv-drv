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

package burmilla

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"shanhu.io/pub/errcode"
)

// DiskUsage contains disk usage data.
type DiskUsage struct {
	Total uint64
	Free  uint64
}

func blockDeviceAndPartition(b *Burmilla) (string, string, error) {
	// TODO(h8liu) : use a better method to recognize platform.
	config, err := ConfigExport(b)
	if err != nil {
		return "", "", errcode.Annotate(err, "read os config")
	}
	docn := isOnDigitalOcean(config)

	switch arch := runtime.GOARCH; arch {
	case "amd64": // NUC
		if docn {
			return "/sys/block/vda", "/dev/vda1", nil
		}
		return "/sys/block/sda", "/dev/sda2", nil
	case "arm64": // RPI
		return "/sys/block/mmcblk0", "/dev/mmcblk0p2", nil
	default:
		return "", "", errcode.Internalf("unknown arch %q", arch)
	}
}

// QueryDiskUsage queries the disk's usage of the system.
func QueryDiskUsage(b *Burmilla) (*DiskUsage, error) {
	blockDevice, part, err := blockDeviceAndPartition(b)
	if err != nil {
		return nil, errcode.Annotate(err, "get block device and partition")
	}

	out, err := b.ExecOutput([]string{"df", part})
	if err != nil {
		return nil, errcode.Annotate(err, "df partition")
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return nil, errcode.Internalf("unexpected df: %q", string(out))
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return nil, errcode.Internalf("unexpected df line: %q", lines[1])
	}

	// df output is in the format of:
	//   <device> <total blocks> <used blocks> <available blocks> ...
	freeBlocks, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return nil, errcode.Internalf("unexpected free blocks: %q", fields[3])
	}

	// df reports in 1024-byte blocks.
	// Also, df reports capacity from the actual usable partition
	// after all the system data and reserved spaces.
	// So this is the actual bytes left that users can write to.
	freeBytes := 1024 * freeBlocks

	out, err = b.ExecOutput(
		[]string{"cat", filepath.Join(blockDevice, "size")})
	if err != nil {
		return nil, errcode.Annotate(err, "query block device size")
	}
	totalBlocks, err :=
		strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return nil, errcode.Internalf("unexpected total blocks: %q", out)
	}
	// totalBlocks are in 512-byte blocks.
	totalBytes := 512 * totalBlocks

	return &DiskUsage{
		Total: totalBytes,
		Free:  freeBytes,
	}, nil
}
