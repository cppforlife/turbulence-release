package controllers

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/turbulence/agentreqs"
	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/scheduledinc"
)

type FactoryRepos interface {
	IncidentsRepo() incident.Repo
	ScheduledIncidentsRepo() scheduledinc.Repo
	AgentRequestsRepo() agentreqs.Repo
}

type Factory struct {
	HomeController               HomeController
	IncidentsController          IncidentsController
	ScheduledIncidentsController ScheduledIncidentsController
	AgentRequestsController      AgentRequestsController
}

func NewFactory(r FactoryRepos, logger boshlog.Logger) (Factory, error) {
	isRepo := r.IncidentsRepo()
	sisRepo := r.ScheduledIncidentsRepo()
	arRepo := r.AgentRequestsRepo()

	factory := Factory{
		HomeController:               NewHomeController(isRepo, sisRepo, logger),
		IncidentsController:          NewIncidentsController(isRepo, logger),
		ScheduledIncidentsController: NewScheduledIncidentsController(sisRepo, logger),
		AgentRequestsController:      NewAgentRequestsController(arRepo, logger),
	}

	return factory, nil
}
