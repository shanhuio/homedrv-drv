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

type task interface {
	run() error
}

type taskEntry struct {
	name string
	task task
	done chan error
}

type taskLoop struct {
	tasks chan *taskEntry
}

func newTaskLoop() *taskLoop {
	return &taskLoop{
		tasks: make(chan *taskEntry, 10),
	}
}

func (l *taskLoop) run(name string, t task) error {
	entry := &taskEntry{
		name: name,
		task: t,
		done: make(chan error),
	}
	l.tasks <- entry
	return <-entry.done
}

func (l *taskLoop) bg() {
	for t := range l.tasks {
		t.done <- t.task.run()
	}
}
