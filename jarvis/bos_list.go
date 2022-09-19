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

	"shanhu.io/misc/errcode"
)

type bosListEntry struct {
	name      string
	tags      []string
	local     bool
	remote    bool
	latest    bool
	running   bool
	available bool
}

func parseBosListEntry(line string) (*bosListEntry, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, errcode.InvalidArgf("empty line")
	}

	entry := &bosListEntry{name: fields[0]}
	entry.tags = append(entry.tags, fields[1:]...)
	for _, tag := range entry.tags {
		switch tag {
		case "local":
			entry.local = true
		case "remote":
			entry.remote = true
		case "running":
			entry.running = true
		case "available":
			entry.available = true
		case "latest":
			entry.latest = true
		}
	}
	return entry, nil
}

type bosList struct {
	entries []*bosListEntry
}

func (ls *bosList) find(v string) *bosListEntry {
	for _, entry := range ls.entries {
		if entry.name == v {
			return entry
		}
	}
	return nil
}

func (ls *bosList) running() *bosListEntry {
	for _, entry := range ls.entries {
		if entry.running {
			return entry
		}
	}
	return nil
}

func parseOSList(lines []string) (*bosList, error) {
	list := new(bosList)
	for _, line := range lines {
		entry, err := parseBosListEntry(line)
		if err != nil {
			return nil, errcode.Annotatef(err, "parse line: %q", line)
		}
		list.entries = append(list.entries, entry)
	}

	return list, nil
}
