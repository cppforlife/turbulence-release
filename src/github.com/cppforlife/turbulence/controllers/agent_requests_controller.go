package controllers

import (
	"encoding/json"
	"net/http"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	mart "github.com/go-martini/martini"
	martrend "github.com/martini-contrib/render"

	"github.com/cppforlife/turbulence/agentreqs"
)

type AgentRequestsController struct {
	agentRequestsRepo agentreqs.Repo

	logTag string
	logger boshlog.Logger
}

func NewAgentRequestsController(
	agentRequestsRepo agentreqs.Repo,
	logger boshlog.Logger,
) AgentRequestsController {
	return AgentRequestsController{
		agentRequestsRepo: agentRequestsRepo,

		logTag: "AgentRequestsController",
		logger: logger,
	}
}

func (c AgentRequestsController) APIConsume(req *http.Request, r martrend.Render, params mart.Params) {
	// agentID := req.URL.Query().Get("agent_id") todo use query string

	tasks, err := c.agentRequestsRepo.Consume(params["id"])
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, tasks)
}

func (c AgentRequestsController) APIUpdate(req *http.Request, r martrend.Render, params mart.Params) {
	var taskReq agentreqs.TaskReq

	err := json.NewDecoder(req.Body).Decode(&taskReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	err = c.agentRequestsRepo.Update(params["id"], taskReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, nil)
}
