package client

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/tasks"
)

type IncidentImpl struct {
	client Client
	id     string
	resp   incident.IncidentResp
}

func (i *IncidentImpl) ID() string { return i.id }

func (i *IncidentImpl) Wait() error {
	for {
		if len(i.resp.ExecutionCompletedAt) > 0 {
			return nil
		}

		var err error

		i.resp, err = i.client.GetIncident(i.id)
		if err != nil {
			return err
		}
	}
}

func (i *IncidentImpl) EventsOfType(example tasks.Options) []reporter.EventResp {
	var events []reporter.EventResp

	for _, ev := range i.resp.Events {
		if ev.Type == tasks.OptionsType(example) {
			events = append(events, ev)
		}
	}

	return events
}

func (i *IncidentImpl) HasEventErrors() bool {
	return i.resp.HasEventErrors()
}

func (c Client) GetIncident(id string) (incident.IncidentResp, error) {
	var resp incident.IncidentResp

	err := c.clientRequest.Get(fmt.Sprintf("/api/v1/incidents/%s", id), &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Getting incident")
	}

	return resp, nil
}
