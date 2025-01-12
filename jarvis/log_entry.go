// Copyright (C) 2023  Shanhu Tech Inc.
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
	"encoding/json"
	"time"

	"shanhu.io/g/rand"
)

// LogEntry is a log entry for jarvis to display.
type LogEntry struct {
	K    string
	T    int64
	User string `json:",omitempty"`
	Text string `json:",omitempty"`
	Type string `json:",omitempty"`
	V    []byte `json:",omitempty"`

	// The following fields are only used in Javascript.
	TSec int64  `json:",omitempty"`
	VStr string `json:",omitempty"`
}

func newLogEntryAt(t time.Time, user, text string) *LogEntry {
	t = t.UTC()
	k := t.Format(time.RFC3339Nano) + "-" + rand.Letters(6)
	return &LogEntry{
		K:    k,
		T:    t.UnixNano(),
		User: user,
		Text: text,
	}
}

func newLogEntry(user, text string) *LogEntry {
	return newLogEntryAt(time.Now(), user, text)
}

func (e *LogEntry) setJSONValue(typ string, v interface{}) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	e.Type = typ
	e.V = bs
	return nil
}

const (
	logTypeLoginAttempt   = "loginAttempt"
	logTypeTwoFactorEvent = "twoFactorEvent"
	logTypeChangePassword = "changePassword"
)
