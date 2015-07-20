package main

import (
	"encoding/json"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/turbulence/agentreqs"
)

type Client struct {
	clientRequest clientRequest
}

func NewClient(endpoint string, httpClient boshhttp.HTTPClient, logger boshlog.Logger) Client {
	clientRequest := clientRequest{
		endpoint:   endpoint,
		httpClient: httpClient,
		logger:     logger,
	}

	return Client{clientRequest: clientRequest}
}

func (c Client) FetchTasks(agentID string) ([]agentreqs.Task, error) {
	var resp []agentreqs.Task

	// todo use query string
	// todo rename to agent_tasks
	path := fmt.Sprintf("/api/v1/agents/%s/tasks", agentID)

	err := c.clientRequest.Post(path, nil, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Fetching tasks '%s'", agentID)
	}

	return resp, nil
}

func (c Client) UpdateTask(taskID string, err error) error {
	var resp interface{}

	path := fmt.Sprintf("/api/v1/agent_tasks/%s", taskID)

	taskReq := agentreqs.TaskReq{}

	if err != nil {
		taskReq.Error = err.Error()
	}

	bytes, err := json.Marshal(taskReq)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshalling task")
	}

	err = c.clientRequest.Post(path, bytes, &resp)
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating task '%s'", taskID)
	}

	return nil
}
