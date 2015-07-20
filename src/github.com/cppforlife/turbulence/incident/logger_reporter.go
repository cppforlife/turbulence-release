package incident

import (
	"fmt"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type LoggerReporter struct {
	logTag string
	logger boshlog.Logger
}

func NewLoggerReporter(logger boshlog.Logger) LoggerReporter {
	return LoggerReporter{
		logTag: "LoggerReporter",
		logger: logger,
	}
}

func (r LoggerReporter) ReportIncidentExecutionStart(i Incident) {
	r.logger.Debug(r.logTag, r.incidentDesc("started", i))
}

func (r LoggerReporter) ReportIncidentExecutionCompletion(i Incident) {
	errorStr := ""

	incidentErr := i.Events.FirstError()
	if incidentErr != nil {
		errorStr = incidentErr.Error()
	}

	r.logger.Debug(r.logTag, "%s error='%s'", r.incidentDesc("completed", i), errorStr)
}

func (r LoggerReporter) ReportEventExecutionStart(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	r.logger.Debug(r.logTag, r.eventDesc("started", e))
}

func (r LoggerReporter) ReportEventExecutionCompletion(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	errorStr := ""

	if e.Error != nil {
		errorStr = e.Error.Error()
	}

	r.logger.Debug(r.logTag, "%s error='%s'", r.eventDesc("completed", e), errorStr)
}

func (r LoggerReporter) incidentDesc(prefix string, i Incident) string {
	return fmt.Sprintf("%s incident='%s' types='%s'",
		prefix, i.ID, strings.Join(i.TaskTypes(), ","))
}

func (r LoggerReporter) eventDesc(prefix string, e Event) string {
	return fmt.Sprintf("%s event='%s' type='%s' deployment='%s' job='%s' index='%d'",
		prefix, e.ID, e.Type, e.DeploymentName, e.JobName, *e.JobIndex)
}
