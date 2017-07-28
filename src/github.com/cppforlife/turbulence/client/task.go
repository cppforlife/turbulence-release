package client

import (
	"errors"
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/tasks"
)

type TaskImpl struct {
	client     Client
	incidentID string
	id         string
}

func (t TaskImpl) Stop() {
	err := t.client.StopTask(t.id)
	panicIfErr(err, "stop task")
}

func (t TaskImpl) Instance() Instance {
	resp := t.fetch()
	return Instance{
		ID:         resp.Instance.ID,
		Group:      resp.Instance.Group,
		Deployment: resp.Instance.Deployment,
		AZ:         resp.Instance.AZ,
	}
}

func (t TaskImpl) Error() string {
	return t.fetch().Error
}

func (t TaskImpl) ExecutionStartedAt() time.Time {
	t1, err := time.Parse(time.RFC3339, t.fetch().ExecutionStartedAt)
	panicIfErr(err, "parse incident's execution start time")

	return t1
}

func (t TaskImpl) ExecutionCompletedAt() *time.Time {
	resp := t.fetch()

	if len(resp.ExecutionCompletedAt) == 0 {
		return nil
	}

	t1, err := time.Parse(time.RFC3339, resp.ExecutionCompletedAt)
	panicIfErr(err, "parse incident's execution completion time")

	return &t1
}

func (t TaskImpl) fetch() reporter.EventResponse {
	incidentResp, err := t.client.GetIncident(t.incidentID)
	panicIfErr(err, "fetch incident response")

	for _, ev := range incidentResp.Events {
		if ev.ID == t.id {
			return ev
		}
	}

	panicIfErr(errors.New("Not found"), "find event associated with task")

	return reporter.EventResponse{} // unreachable
}

func (c Client) StopTask(id string) error {
	var resp interface{}

	path := fmt.Sprintf("/api/v1/agent_tasks/%s/state", id)
	req := tasks.StateRequest{Stop: true}

	err := c.clientRequest.Post(path, req, &resp)
	if err != nil {
		return bosherr.WrapErrorf(err, "Stopping task")
	}

	return nil
}
