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

const (
	keySessionHMAC = "session.hmac"

	// Init passwords.
	keyJarvisPass         = "jarvis.pass"
	keyPostgresPass       = "postgress.pass"
	keyNextcloudDBPass    = "nextcloud-db.pass"
	keyNextcloudAdminPass = "nextcloud-admin.pass"
	keyRedisPass          = "redis.pass"

	keyMainDomain          = "main.domain"
	keyNextcloudDomain     = "nextcloud.domain"
	keyNextcloudDomains    = "nextcloud.domains"
	keyNextcloudDataMount  = "nextcloud.data-mount"
	keyFabricsServerDomain = "fabrics-server.domain"
	keyCustomSubs          = "custom.subs"

	keyBuild         = "build"
	keyBuildUpdating = "build-updating"
	keyManualBuild   = "manual-build"

	keyNextcloud18Fixed = "nextcloud-18-fixed"
	keyNextcloud19Fixed = "nextcloud-19-fixed"
	keyNextcloud20Fixed = "nextcloud-20-fixed"
	keyNextcloud21Fixed = "nextcloud-21-fixed"

	keyIdentity = "identity"

	keyAppsState = "apps.state"
)
