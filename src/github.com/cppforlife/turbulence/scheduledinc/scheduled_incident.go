package scheduledinc

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/turbulence/incident"
)

type ScheduledIncident struct {
	updateFunc    func(ScheduledIncident) error
	incidentsRepo incident.Repo
	logger        boshlog.Logger

	ID string

	Schedule string

	Incident incident.Request
}

func (si ScheduledIncident) Execute() error {
	_, err := si.incidentsRepo.Create(si.Incident)
	if err != nil {
		return bosherr.WrapErrorf(err,
			"Creating incident based on scheduled incident ID '%s'", si.ID)
	}

	return nil
}
