package tasks

// todo should not be an agent task
type KillOptions struct {
	Type string
}

func (KillOptions) _private() {}
