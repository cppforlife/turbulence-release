package main

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"github.com/cppforlife/turbulence/agentreqs"
	"github.com/cppforlife/turbulence/director"
	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/scheduledinc"
)

type Repos struct {
	incidentsRepo          incident.Repo
	scheduledIncidentsRepo scheduledinc.Repo
	agentRequestsRepo      agentreqs.Repo
}

func NewRepos(
	uuidGen boshuuid.Generator,
	reporter reporter.Reporter,
	director director.Director,
	incidentNotifier incident.RepoNotifier,
	scheduledIncidentNotifier scheduledinc.RepoNotifier,
	logger boshlog.Logger,
) (Repos, error) {
	agentRequestRepo := agentreqs.NewRepo(logger)

	incidentsRepo := incident.NewRepo(
		uuidGen,
		incidentNotifier,
		reporter,
		director,
		agentRequestRepo,
		logger,
	)

	scheduledIncidentsRepo := scheduledinc.NewRepo(
		uuidGen,
		scheduledIncidentNotifier,
		incidentsRepo,
		logger,
	)

	return Repos{incidentsRepo, scheduledIncidentsRepo, agentRequestRepo}, nil
}

func (r Repos) IncidentsRepo() incident.Repo              { return r.incidentsRepo }
func (r Repos) ScheduledIncidentsRepo() scheduledinc.Repo { return r.scheduledIncidentsRepo }
func (r Repos) AgentRequestsRepo() agentreqs.Repo         { return r.agentRequestsRepo }
