package tasks

type Task struct {
	ID string

	Optionss OptionsSlice
}

type OptionsSlice []Options

type Options interface {
	_private()
}

type Repo interface {
	QueueAndWait(string, []Task) error
	Consume(string) ([]Task, error)

	Wait(string) (TaskReq, error)
	Update(string, TaskReq) error
}
