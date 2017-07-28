package client

import (
	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/tasks"
)

type Turbulence interface {
	CreateIncident(incident.IncidentReq) (Incident, error)
}

type Incident interface {
	ID() string
	Wait() error // todo add timeout?

	EventsOfType(tasks.TaskOptions) []reporter.EventResp
	HasEventErrors() bool
}
