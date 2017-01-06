package selector

type Instance interface {
	ID() string
	Group() string
	Deployment() string
	AZ() string
	HasVM() bool
}

type Selector interface {
	Select([]Instance) ([]Instance, error)
}

type Limitor interface {
	Limit([]string) ([]string, error)
}
