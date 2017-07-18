package client

import (
	"github.com/cppforlife/turbulence/agentreqs"
	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/incident/reporter"
)

type Turbulence interface {
	CreateIncident(incident.IncidentReq) (Incident, error)
}

type Incident interface {
	ID() string
	Wait() error // todo add timeout?

	EventsOfType(agentreqs.TaskOptions) []reporter.EventResp
	HasEventErrors() bool
}
