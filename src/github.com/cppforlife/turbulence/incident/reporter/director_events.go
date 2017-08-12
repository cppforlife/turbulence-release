package reporter

import (
	"fmt"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/turbulence/director"
)

type DirectorEvents struct {
	director director.Director

	logTag string
	logger boshlog.Logger
}

func NewDirectorEvents(director director.Director, logger boshlog.Logger) DirectorEvents {
	return DirectorEvents{director, "incident.reporter.DirectorEvents", logger}
}

func (r DirectorEvents) ReportIncidentExecutionStart(i Incident) {
	err := r.director.SubmitEvent(director.EventOpts{
		Action:     "start",
		ObjectType: "turbulence-incident",
		ObjectName: i.ID(),
		Context: map[string]interface{}{
			"summary": strings.Join(i.TaskTypes(), ","),
		},
	})
	r.logErr(err)
}

func (r DirectorEvents) ReportIncidentExecutionCompletion(i Incident) {
	errorStr := ""

	incidentErr := i.Events().FirstError()
	if incidentErr != nil {
		errorStr = incidentErr.Error()
	}

	err := r.director.SubmitEvent(director.EventOpts{
		Action:     "end",
		ObjectType: "turbulence-incident",
		ObjectName: i.ID(),
		Error:      errorStr,
	})
	r.logErr(err)
}

func (r DirectorEvents) ReportEventExecutionStart(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	err := r.director.SubmitEvent(director.EventOpts{
		Action:     "start",
		ObjectType: "turbulence-event",
		ObjectName: e.ID,
		Deployment: e.Instance.Deployment,
		Instance:   fmt.Sprintf("%s/%s", e.Instance.Group, e.Instance.ID),
		Context: map[string]interface{}{
			"type":        e.Type,
			"incident_id": incidentID,
		},
	})
	r.logErr(err)
}

func (r DirectorEvents) ReportEventExecutionCompletion(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	errorStr := ""

	if e.Error != nil {
		errorStr = e.Error.Error()
	}

	err := r.director.SubmitEvent(director.EventOpts{
		Action:     "end",
		ObjectType: "turbulence-event",
		ObjectName: e.ID,
		Deployment: e.Instance.Deployment,
		Instance:   fmt.Sprintf("%s/%s", e.Instance.Group, e.Instance.ID),
		Context:    map[string]interface{}{"incident_id": incidentID},
		Error:      errorStr,
	})
	r.logErr(err)
}

func (r DirectorEvents) logErr(err error) {
	if err != nil {
		r.logger.Error(r.logTag, "Failed submitting event: %s", err)
	}
}
