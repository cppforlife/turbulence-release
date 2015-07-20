package incident

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Worker struct {
	logTag string
	logger boshlog.Logger
}

func NewWorker(logger boshlog.Logger) Worker {
	return Worker{logTag: "Worker", logger: logger}
}

func (w Worker) IncidentWasCreated(incident Incident) {
	err := incident.Execute()
	if err != nil {
		w.logger.Error(w.logTag, "Failed to execute incident: %s", err.Error())
	}
}
