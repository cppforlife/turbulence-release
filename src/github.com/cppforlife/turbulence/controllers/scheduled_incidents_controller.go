package controllers

import (
	"encoding/json"
	"net/http"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	mart "github.com/go-martini/martini"
	martrend "github.com/martini-contrib/render"

	"github.com/cppforlife/turbulence/scheduledinc"
)

type ScheduledIncidentsController struct {
	repo scheduledinc.Repo

	indexTmpl string
	showTmpl  string
	errorTmpl string

	logTag string
	logger boshlog.Logger
}

func NewScheduledIncidentsController(
	repo scheduledinc.Repo,
	logger boshlog.Logger,
) ScheduledIncidentsController {
	return ScheduledIncidentsController{
		repo: repo,

		indexTmpl: "scheduled_incidents/index",
		showTmpl:  "scheduled_incidents/show",
		errorTmpl: "error",

		logTag: "ScheduledIncidentsController",
		logger: logger,
	}
}

type ScheduledIncidentsPage struct {
	ScheduledIncidents []scheduledinc.ScheduledIncidentResp
}

type ScheduledIncidentPage struct {
	ScheduledIncident scheduledinc.ScheduledIncidentResp
}

func (c ScheduledIncidentsController) Index(r martrend.Render) {
	sis, err := c.repo.ListAll()
	if err != nil {
		r.HTML(500, c.errorTmpl, err)
		return
	}

	r.HTML(200, c.indexTmpl, ScheduledIncidentsPage{scheduledinc.NewScheduledIncidentsResp(sis)})
}

func (c ScheduledIncidentsController) APIIndex(r martrend.Render) {
	sis, err := c.repo.ListAll()
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, scheduledinc.NewScheduledIncidentsResp(sis))
}

func (c ScheduledIncidentsController) APICreate(req *http.Request, r martrend.Render) {
	var siReq scheduledinc.ScheduledRequest

	err := json.NewDecoder(req.Body).Decode(&siReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	si, err := c.repo.Create(siReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, scheduledinc.NewScheduledIncidentResp(si))
}

func (c ScheduledIncidentsController) Read(r martrend.Render, params mart.Params) {
	si, err := c.repo.Read(params["id"])
	if err != nil {
		code := 500
		if _, ok := err.(scheduledinc.ScheduledIncidentNotFoundError); ok {
			code = 404
		}

		r.HTML(code, c.errorTmpl, err)
		return
	}

	r.HTML(200, c.showTmpl, ScheduledIncidentPage{scheduledinc.NewScheduledIncidentResp(si)})
}

func (c ScheduledIncidentsController) APIRead(r martrend.Render, params mart.Params) {
	si, err := c.repo.Read(params["id"])
	if err != nil {
		code := 500
		if _, ok := err.(scheduledinc.ScheduledIncidentNotFoundError); ok {
			code = 404
		}

		r.JSON(code, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, scheduledinc.NewScheduledIncidentResp(si))
}

func (c ScheduledIncidentsController) APIDelete(r martrend.Render, params mart.Params) {
	err := c.repo.Delete(params["id"])
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, map[string]string{})
}
