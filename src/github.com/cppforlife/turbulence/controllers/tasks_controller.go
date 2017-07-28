package controllers

import (
	"encoding/json"
	"net/http"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	mart "github.com/go-martini/martini"
	martrend "github.com/martini-contrib/render"

	"github.com/cppforlife/turbulence/tasks"
)

type TasksController struct {
	tasksRepo tasks.Repo

	logTag string
	logger boshlog.Logger
}

func NewTasksController(
	tasksRepo tasks.Repo,
	logger boshlog.Logger,
) TasksController {
	return TasksController{
		tasksRepo: tasksRepo,

		logTag: "TasksController",
		logger: logger,
	}
}

func (c TasksController) APIConsume(req *http.Request, r martrend.Render, params mart.Params) {
	// agentID := req.URL.Query().Get("agent_id") todo use query string

	tasks, err := c.tasksRepo.Consume(params["id"])
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, tasks)
}

func (c TasksController) APIUpdate(req *http.Request, r martrend.Render, params mart.Params) {
	var taskReq tasks.Request

	err := json.NewDecoder(req.Body).Decode(&taskReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	err = c.tasksRepo.Update(params["id"], taskReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, nil)
}
