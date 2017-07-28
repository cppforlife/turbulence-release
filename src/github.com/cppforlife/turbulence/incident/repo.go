package incident

import (
	"fmt"
	"sync"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"github.com/cppforlife/turbulence/director"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/tasks"
)

type IncidentNotFoundError struct {
	ID string
}

func (e IncidentNotFoundError) Error() string {
	return fmt.Sprintf("Incident '%s' does not exist", e.ID)
}

type repo struct {
	uuidGen   boshuuid.Generator
	notifier  RepoNotifier
	reporter  reporter.Reporter
	director  director.Director
	tasksRepo tasks.Repo

	incidents     []Incident
	incidentsLock sync.RWMutex

	logger boshlog.Logger
}

func NewRepo(
	uuidGen boshuuid.Generator,
	notifier RepoNotifier,
	reporter reporter.Reporter,
	director director.Director,
	tasksRepo tasks.Repo,
	logger boshlog.Logger,
) Repo {
	return &repo{
		uuidGen:   uuidGen,
		notifier:  notifier,
		reporter:  reporter,
		director:  director,
		tasksRepo: tasksRepo,
		logger:    logger,
	}
}

func (r *repo) ListAll() ([]Incident, error) {
	var reversed []Incident

	for i := len(r.incidents) - 1; i >= 0; i-- {
		reversed = append(reversed, r.incidents[i])
	}

	return reversed, nil
}

func (r *repo) Create(req Request) (Incident, error) {
	id, err := r.uuidGen.Generate()
	if err != nil {
		return Incident{}, bosherr.WrapError(err, "Generating incident ID")
	}

	incident := Incident{
		director:   r.director,
		reporter:   r.reporter,
		tasksRepo:  r.tasksRepo,
		updateFunc: r.update,

		id: id,

		Tasks:    req.Tasks,
		Selector: req.Selector,

		events: reporter.NewEvents(r.uuidGen, r.reporter, id, r.logger),

		logTag: "incident.Incident",
		logger: r.logger,
	}

	r.incidentsLock.Lock()
	r.incidents = append(r.incidents, incident)
	r.incidentsLock.Unlock()

	// notified after incidents were unlocked
	go r.notifier.IncidentWasCreated(incident)

	return incident, nil
}

func (r *repo) Read(id string) (Incident, error) {
	r.incidentsLock.Lock()
	defer r.incidentsLock.Unlock()

	for _, incident := range r.incidents {
		if incident.ID() == id {
			return incident, nil
		}
	}

	return Incident{}, IncidentNotFoundError{ID: id}
}

func (r *repo) update(updatedIncident Incident) error {
	r.incidentsLock.Lock()
	defer r.incidentsLock.Unlock()

	for i, incident := range r.incidents {
		if incident.ID() == updatedIncident.ID() {
			r.incidents[i] = updatedIncident
			break
		}
	}

	return nil
}
