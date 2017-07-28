package controllers

import (
	"encoding/json"
	"net/http"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	mart "github.com/go-martini/martini"
	martrend "github.com/martini-contrib/render"

	"github.com/cppforlife/turbulence/incident"
)

type IncidentsController struct {
	incidentsRepo incident.Repo

	indexTmpl string
	showTmpl  string
	errorTmpl string

	logTag string
	logger boshlog.Logger
}

func NewIncidentsController(
	incidentsRepo incident.Repo,
	logger boshlog.Logger,
) IncidentsController {
	return IncidentsController{
		incidentsRepo: incidentsRepo,

		indexTmpl: "incidents/index",
		showTmpl:  "incidents/show",
		errorTmpl: "error",

		logTag: "IncidentsController",
		logger: logger,
	}
}

type IncidentsPage struct {
	Incidents []incident.Response
}

type IncidentPage struct {
	Incident incident.Response
}

func (c IncidentsController) Index(req *http.Request, r martrend.Render) {
	incidents, err := c.incidentsRepo.ListAll()
	if err != nil {
		r.HTML(500, c.errorTmpl, err)
		return
	}

	r.HTML(200, c.indexTmpl, IncidentsPage{incident.NewResponses(incidents)})
}

func (c IncidentsController) APIIndex(req *http.Request, r martrend.Render) {
	incidents, err := c.incidentsRepo.ListAll()
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, incident.NewResponses(incidents))
}

func (c IncidentsController) APICreate(req *http.Request, r martrend.Render) {
	var incidentReq incident.Request

	err := json.NewDecoder(req.Body).Decode(&incidentReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	incid, err := c.incidentsRepo.Create(incidentReq)
	if err != nil {
		r.JSON(500, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, incident.NewResponse(incid))
}

func (c IncidentsController) Read(req *http.Request, r martrend.Render, params mart.Params) {
	incid, err := c.incidentsRepo.Read(params["id"])
	if err != nil {
		code := 500
		if _, ok := err.(incident.IncidentNotFoundError); ok {
			code = 404
		}

		r.HTML(code, c.errorTmpl, err)
		return
	}

	r.HTML(200, c.showTmpl, IncidentPage{incident.NewResponse(incid)})
}

func (c IncidentsController) APIRead(req *http.Request, r martrend.Render, params mart.Params) {
	incid, err := c.incidentsRepo.Read(params["id"])
	if err != nil {
		code := 500
		if _, ok := err.(incident.IncidentNotFoundError); ok {
			code = 404
		}

		r.JSON(code, map[string]string{"error": err.Error()})
		return
	}

	r.JSON(200, incident.NewResponse(incid))
}
