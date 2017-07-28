package client

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/tasks"
)

type TaskImpl struct {
	client Client
	id     string
}

func (i *TaskImpl) Stop() {
	err := i.client.StopTask(i.id)
	panicIfErr(err, "stop task")
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
