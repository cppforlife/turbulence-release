package tasks

type NoopOptions struct {
	Type      string
	Stoppable bool
}

func (NoopOptions) _private() {}

type NoopTask struct {
	opts NoopOptions
}

func NewNoopTask(opts NoopOptions) NoopTask {
	return NoopTask{opts}
}

func (t NoopTask) Execute(stopCh chan struct{}) error {
	if t.opts.Stoppable {
		<-stopCh
	}

	return nil
}
