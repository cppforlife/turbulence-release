package client

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/tasks"
)

type IncidentImpl struct {
	client Client
	id     string
	resp   incident.Response
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

func (i *IncidentImpl) EventsOfType(example tasks.Options) []reporter.EventResponse {
	var events []reporter.EventResponse

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

func (i *IncidentImpl) ExecutionStartedAt() time.Time {
	t, _ := time.Parse(time.RFC3339, i.resp.ExecutionStartedAt) // todo err check?
	return t
}

func (i *IncidentImpl) ExecutionCompletedAt() *time.Time {
	if len(i.resp.ExecutionCompletedAt) == 0 {
		return nil
	}
	t, _ := time.Parse(time.RFC3339, i.resp.ExecutionCompletedAt)
	return &t
}

func (c Client) GetIncident(id string) (incident.Response, error) {
	var resp incident.Response

	err := c.clientRequest.Get(fmt.Sprintf("/api/v1/incidents/%s", id), &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Getting incident")
	}

	return resp, nil
}
