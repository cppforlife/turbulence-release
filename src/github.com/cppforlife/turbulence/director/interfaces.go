package director

type Director interface {
	AllInstances() ([]Instance, error)
}

type Instance interface {
	ID() string
	Group() string
	Deployment() string
	AZ() string

	AgentID() string
	HasVM() bool

	DeleteVM() error
}
