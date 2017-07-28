package client

import (
	"time"

	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/tasks"
)

// Turbulence client is designed to be test friendly
// hence it does not return errors but rather panics.
type Turbulence interface {
	CreateIncident(incident.Request) Incident
}

type Incident interface {
	Wait() // todo add timeout?

	TasksOfType(tasks.Options) []Task
	HasTaskErrors() bool

	// ExecutionStartedAt is expected to always return time,
	// unlike ExecutionCompletedAt which may return nil
	// when execution is not yet finished
	ExecutionStartedAt() time.Time
	ExecutionCompletedAt() *time.Time
}

var _ Incident = IncidentImpl{}

type Task interface {
	Stop()

	Instance() Instance
	Error() string

	ExecutionStartedAt() time.Time
	ExecutionCompletedAt() *time.Time
}

var _ Task = TaskImpl{}

type Instance struct {
	ID         string
	Group      string
	Deployment string
	AZ         string
}
