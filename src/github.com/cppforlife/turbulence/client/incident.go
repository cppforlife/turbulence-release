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
}

func (i *IncidentImpl) Wait() {
	for len(i.fetch().ExecutionCompletedAt) == 0 {
		time.Sleep(500 * time.Millisecond)
	}
}

func (i *IncidentImpl) Tasks() []Task {
	var ts []Task

	for _, ev := range i.fetch().Events {
		ts = append(ts, &TaskImpl{i.client, ev.ID})
	}

	return ts
}

func (i *IncidentImpl) TasksOfType(example tasks.Options) []Task {
	var ts []Task

	for _, ev := range i.fetch().Events {
		if ev.Type == tasks.OptionsType(example) {
			ts = append(ts, &TaskImpl{i.client, ev.ID})
		}
	}

	return ts
}

func (i *IncidentImpl) EventsOfType(example tasks.Options) []reporter.EventResponse {
	var events []reporter.EventResponse

	for _, ev := range i.fetch().Events {
		if ev.Type == tasks.OptionsType(example) {
			events = append(events, ev)
		}
	}

	return events
}

func (i *IncidentImpl) HasEventErrors() bool {
	return i.fetch().HasEventErrors()
}

func (i *IncidentImpl) ExecutionStartedAt() time.Time {
	t, err := time.Parse(time.RFC3339, i.fetch().ExecutionStartedAt)
	panicIfErr(err, "parse incident's execution start time")

	return t
}

func (i *IncidentImpl) ExecutionCompletedAt() *time.Time {
	resp := i.fetch()

	if len(resp.ExecutionCompletedAt) == 0 {
		return nil
	}

	t, err := time.Parse(time.RFC3339, resp.ExecutionCompletedAt)
	panicIfErr(err, "parse incident's execution completion time")

	return &t
}

func (i *IncidentImpl) fetch() incident.Response {
	resp, err := i.client.GetIncident(i.id)
	panicIfErr(err, "fetch incident response")

	return resp
}

func (c Client) GetIncident(id string) (incident.Response, error) {
	var resp incident.Response

	err := c.clientRequest.Get(fmt.Sprintf("/api/v1/incidents/%s", id), &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Getting incident")
	}

	return resp, nil
}
