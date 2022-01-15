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
	"fmt"

	"shanhu.io/misc/errcode"
	"shanhu.io/misc/strutil"
)

func cmdOCC(args []string) error {
	flags := cmdFlags.New()
	cflags := newClientFlags(flags)
	args = flags.ParseArgs(args)
	d, err := newClientDrive(cflags)
	if err != nil {
		return err
	}

	nc := newNextcloud(d)
	return nc.occ(args, nil)
}

func cmdNextcloudDomains(args []string) error {
	flags := cmdFlags.New()
	cflags := newClientFlags(flags)
	addDomain := flags.String("add", "", "add a new nextcloud domain")
	removeDomain := flags.String("remove", "", "remove a nextcloud domain")
	flags.ParseArgs(args)

	d, err := newClientDrive(cflags)
	if err != nil {
		return err
	}

	domains, err := nextcloudDomains(d)
	if err != nil {
		return errcode.Annotate(err, "load domains")
	}

	if *addDomain == "" && *removeDomain == "" {
		for _, domain := range domains {
			fmt.Println(domain)
		}
		return nil
	}

	domainSet := strutil.MakeSet(domains)
	if domain := *addDomain; domain != "" {
		if domainSet[domain] {
			return errcode.InvalidArgf("domain %q already exist", domain)
		}
		domainSet[domain] = true
	}
	if domain := *removeDomain; domain != "" {
		if !domainSet[domain] {
			return errcode.InvalidArgf("domain %q not in domain list", domain)
		}
		if len(domainSet) == 1 {
			return errcode.Internalf(
				"cannot remove the the only domain, " +
					"please add a new one first",
			)
		}
		delete(domainSet, domain)
	}

	domains = strutil.SortedList(domainSet)
	return d.settings.Set(keyNextcloudDomains, domains)
}
