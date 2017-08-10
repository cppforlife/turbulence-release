package director

type Director interface {
	AllInstances() ([]Instance, error)
	SubmitEvent(EventOpts) error
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

type EventOpts struct {
	Action     string
	ObjectType string
	ObjectName string
	Deployment string
	Instance   string
	Context    map[string]interface{}
	Error      string
}
