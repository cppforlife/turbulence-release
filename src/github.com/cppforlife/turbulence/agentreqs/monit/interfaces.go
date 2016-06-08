package monit

type Client interface {
	Services() ([]Service, error)
}

type Service struct {
	Name string
	PID  int
}
