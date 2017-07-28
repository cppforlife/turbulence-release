package tasks

type Task struct {
	ID string

	Optionss OptionsSlice // todo shoudl be singular
}

type Options interface {
	_private()
}

func (t Task) Options() Options {
	return t.Optionss[0]
}

type OptionsSlice []Options

type State struct {
	Stop bool
}

type Repo interface {
	QueueAndWait(string, []Task) error
	Consume(string) ([]Task, error)

	Wait(string) (ResultRequest, error)
	Update(string, ResultRequest) error

	FetchState(string) (State, error)
	UpdateState(string, StateRequest) error
}
