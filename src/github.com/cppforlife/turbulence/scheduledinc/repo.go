package scheduledinc

import (
	"fmt"
	"sync"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"github.com/cppforlife/turbulence/incident"
)

type ScheduledIncidentNotFoundError struct {
	ID string
}

func (e ScheduledIncidentNotFoundError) Error() string {
	return fmt.Sprintf("Scheduled incident '%s' does not exist", e.ID)
}

type Repo interface {
	ListAll() ([]ScheduledIncident, error)
	Create(ScheduledRequest) (ScheduledIncident, error)
	Read(string) (ScheduledIncident, error)
	Delete(string) error
}

type RepoNotifier interface {
	ScheduledIncidentWasCreated(ScheduledIncident)
	ScheduledIncidentWasDeleted(ScheduledIncident)
}

type repo struct {
	uuidGen       boshuuid.Generator
	notifier      RepoNotifier
	incidentsRepo incident.Repo

	sis     []ScheduledIncident
	sisLock sync.RWMutex

	logger boshlog.Logger
}

func NewRepo(
	uuidGen boshuuid.Generator,
	notifier RepoNotifier,
	incidentsRepo incident.Repo,
	logger boshlog.Logger,
) Repo {
	return &repo{
		uuidGen:       uuidGen,
		notifier:      notifier,
		incidentsRepo: incidentsRepo,
		logger:        logger,
	}
}

func (r *repo) ListAll() ([]ScheduledIncident, error) {
	// todo lock?
	return r.sis, nil
}

func (r *repo) Create(req ScheduledRequest) (ScheduledIncident, error) {
	uuid, err := r.uuidGen.Generate()
	if err != nil {
		return ScheduledIncident{}, bosherr.WrapError(err, "Generating scheduled incident ID")
	}

	scheduledIncident := ScheduledIncident{
		updateFunc:    r.update,
		incidentsRepo: r.incidentsRepo,
		logger:        r.logger,

		ID: uuid,

		Schedule: req.Schedule,
		Incident: req.Incident,
	}

	r.sisLock.Lock()
	r.sis = append(r.sis, scheduledIncident)
	r.sisLock.Unlock()

	// notified after scheduled incidents were unlocked
	go r.notifier.ScheduledIncidentWasCreated(scheduledIncident)

	return scheduledIncident, nil
}

func (r *repo) Read(id string) (ScheduledIncident, error) {
	r.sisLock.Lock()
	defer r.sisLock.Unlock()

	for _, si := range r.sis {
		if si.ID == id {
			return si, nil
		}
	}

	return ScheduledIncident{}, ScheduledIncidentNotFoundError{ID: id}
}

func (r *repo) Delete(id string) error {
	var deletedSi ScheduledIncident

	r.sisLock.Lock()

	for i, si := range r.sis {
		if si.ID == id {
			deletedSi = si
			r.sis = append(r.sis[:i], r.sis[i+1:]...)
			break
		}
	}

	r.sisLock.Unlock()

	// notified after scheduled incidents were unlocked
	if len(deletedSi.ID) > 0 {
		go r.notifier.ScheduledIncidentWasDeleted(deletedSi)
	}

	return nil
}

func (r *repo) update(updated ScheduledIncident) error {
	r.sisLock.Lock()
	defer r.sisLock.Unlock()

	for i, incident := range r.sis {
		if incident.ID == updated.ID {
			r.sis[i] = updated
			break
		}
	}

	return nil
}
