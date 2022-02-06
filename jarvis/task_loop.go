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
