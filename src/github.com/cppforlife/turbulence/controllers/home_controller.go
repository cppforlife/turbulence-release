package controllers

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	martrend "github.com/martini-contrib/render"

	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/scheduledinc"
)

type HomeController struct {
	incidentsRepo          incident.Repo
	scheduledIncidentsRepo scheduledinc.Repo

	homeTmpl  string
	errorTmpl string

	logTag string
	logger boshlog.Logger
}

func NewHomeController(
	incidentsRepo incident.Repo,
	scheduledIncidentsRepo scheduledinc.Repo,
	logger boshlog.Logger,
) HomeController {
	return HomeController{
		incidentsRepo:          incidentsRepo,
		scheduledIncidentsRepo: scheduledIncidentsRepo,

		homeTmpl:  "home/home",
		errorTmpl: "error",

		logTag: "HomeController",
		logger: logger,
	}
}

type HomePage struct {
	Incidents          incident.IncidentsResp
	ScheduledIncidents scheduledinc.ScheduledIncidentsResp
}

func (c HomeController) Home(r martrend.Render) {
	is, err := c.incidentsRepo.ListAll()
	if err != nil {
		r.HTML(500, c.errorTmpl, err)
		return
	}

	sis, err := c.scheduledIncidentsRepo.ListAll()
	if err != nil {
		r.HTML(500, c.errorTmpl, err)
		return
	}

	page := HomePage{
		Incidents:          incident.NewResponses(is),
		ScheduledIncidents: scheduledinc.NewScheduledIncidentsResp(sis),
	}

	r.HTML(200, c.homeTmpl, page)
}
