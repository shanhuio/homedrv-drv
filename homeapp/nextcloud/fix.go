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

package nextcloud

import (
	"io"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/g/semver"
	"shanhu.io/g/settings"
)

func fix(cont *dock.Cont, s settings.Settings) error {
	version, err := readTrueVersion(cont)
	if err != nil {
		return errcode.Annotate(err, "get version")
	}
	major, err := semver.Major(version)
	if err != nil {
		return errcode.Add(errcode.Internal, err)
	}
	return fixVersion(cont, s, major)
}

func fixVersion(cont *dock.Cont, s settings.Settings, major int) error {
	if major >= 21 {
		// For version 21+, this needs to be executed every time a new
		// docker is installed.
		if err := aptUpdate(cont, io.Discard); err != nil {
			return errcode.Annotate(err, "apt update for nc21")
		}

		pkgs := []string{
			"libmagickcore-6.q16-6-extra",
			"smbclient",
			"libsmbclient-dev",
		}
		if err := aptInstall(cont, pkgs, io.Discard); err != nil {
			return errcode.Annotate(err, "install additional packages")
		}

		if err := enableSMB(cont, io.Discard); err != nil {
			return errcode.Annotate(err, "enable SMB")
		}
	}

	if err := setCronMode(cont); err != nil {
		return errcode.Annotate(err, "set cron mode")
	}

	// The following fixes might be also needed in minor upgrades.
	for _, cmd := range []string{
		"db:add-missing-indices",
		"db:convert-filecache-bigint",
	} {
		if _, err := occOutput(cont, []string{cmd, "-n"}); err != nil {
			return errcode.Annotate(err, cmd)
		}
	}

	k := fixKey(major)
	if k == "" {
		return nil
	}
	ok, err := s.Has(k)
	if err != nil {
		return errcode.Annotatef(err, "check fixed flag v%d", major)
	}
	if ok {
		return nil
	}

	for _, cmd := range []string{
		"db:add-missing-columns",
		"db:add-missing-primary-keys",
	} {
		if _, err := occOutput(cont, []string{cmd, "-n"}); err != nil {
			return errcode.Annotate(err, cmd)
		}
	}

	if err := s.Set(k, true); err != nil {
		return errcode.Annotatef(err, "set fixed flag v%d", major)
	}
	return nil
}
