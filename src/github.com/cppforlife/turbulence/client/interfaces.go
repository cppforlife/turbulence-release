package client

import (
	"time"

	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/tasks"
)

// Turbulence client is designed to be test friendly
// hence it does not return errors but rather panics.
type Turbulence interface {
	CreateIncident(incident.Request) Incident
}

type Incident interface {
	Wait() // todo add timeout?

	Tasks() []Task
	TasksOfType(tasks.Options) []Task

	// EventsOfType returns list events that match particular options type
	// Example: incident.EventsOfType(tasks.KillOptions{})
	EventsOfType(tasks.Options) []reporter.EventResponse
	HasEventErrors() bool

	// ExecutionStartedAt is expected to always return time,
	// unlike ExecutionCompletedAt which may return nil
	// when execution is not yet finished
	ExecutionStartedAt() time.Time
	ExecutionCompletedAt() *time.Time
}

var _ Incident = &IncidentImpl{}

type Task interface {
	Stop()
}

var _ Task = &TaskImpl{}
