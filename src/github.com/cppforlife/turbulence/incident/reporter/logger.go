package reporter

import (
	"fmt"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Logger struct {
	logTag string
	logger boshlog.Logger
}

func NewLogger(logger boshlog.Logger) Logger {
	return Logger{"incident.reporter.Logger", logger}
}

func (r Logger) ReportIncidentExecutionStart(i Incident) {
	r.logger.Debug(r.logTag, r.incidentDesc("started", i))
}

func (r Logger) ReportIncidentExecutionCompletion(i Incident) {
	errorStr := ""

	incidentErr := i.Events().FirstError()
	if incidentErr != nil {
		errorStr = incidentErr.Error()
	}

	r.logger.Debug(r.logTag, "%s error='%s'", r.incidentDesc("completed", i), errorStr)
}

func (r Logger) ReportEventExecutionStart(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	r.logger.Debug(r.logTag, r.eventDesc("started", e))
}

func (r Logger) ReportEventExecutionCompletion(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	errorStr := ""

	if e.Error != nil {
		errorStr = e.Error.Error()
	}

	r.logger.Debug(r.logTag, "%s error='%s'", r.eventDesc("completed", e), errorStr)
}

func (r Logger) incidentDesc(prefix string, i Incident) string {
	return fmt.Sprintf("%s incident='%s' types='%s'", prefix, i.ID, strings.Join(i.TaskTypes(), ","))
}

func (r Logger) eventDesc(prefix string, e Event) string {
	return fmt.Sprintf("%s event='%s' type='%s' deployment='%s' instance='%s/%s'", prefix, e.ID, e.Type, e.Instance.Deployment, e.Instance.Group, e.Instance.ID)
}
