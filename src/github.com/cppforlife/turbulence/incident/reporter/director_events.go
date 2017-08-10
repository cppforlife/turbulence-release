package reporter

import (
	"fmt"
	"strings"

	"github.com/cppforlife/turbulence/director"
)

type DirectorEvents struct {
	director director.Director
}

func NewDirectorEvents(director director.Director) DirectorEvents {
	return DirectorEvents{director}
}

func (r DirectorEvents) ReportIncidentExecutionStart(i Incident) {
	r.director.SubmitEvent(director.EventOpts{
		Action:     "start",
		ObjectType: "turbulence-incident",
		ObjectName: i.ID(),
		Context: map[string]interface{}{
			"summary": strings.Join(i.TaskTypes(), ","),
		},
	})
}

func (r DirectorEvents) ReportIncidentExecutionCompletion(i Incident) {
	errorStr := ""

	incidentErr := i.Events().FirstError()
	if incidentErr != nil {
		errorStr = incidentErr.Error()
	}

	r.director.SubmitEvent(director.EventOpts{
		Action:     "end",
		ObjectType: "turbulence-incident",
		ObjectName: i.ID(),
		Error:      errorStr,
	})
}

func (r DirectorEvents) ReportEventExecutionStart(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	r.director.SubmitEvent(director.EventOpts{
		Action:     "start",
		ObjectType: "turbulence-event",
		ObjectName: e.ID,
		Deployment: e.Instance.Deployment,
		Instance:   fmt.Sprintf("%s/%s", e.Instance.Group, e.Instance.ID),
		Context:    map[string]interface{}{"incident_id": incidentID},
	})
}

func (r DirectorEvents) ReportEventExecutionCompletion(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	errorStr := ""

	if e.Error != nil {
		errorStr = e.Error.Error()
	}

	r.director.SubmitEvent(director.EventOpts{
		Action:     "end",
		ObjectType: "turbulence-event",
		ObjectName: e.ID,
		Deployment: e.Instance.Deployment,
		Instance:   fmt.Sprintf("%s/%s", e.Instance.Group, e.Instance.ID),
		Context:    map[string]interface{}{"incident_id": incidentID},
		Error:      errorStr,
	})
}

func (r DirectorEvents) eventDesc(prefix string, e Event) string {
	return fmt.Sprintf("%s event='%s' type='%s' deployment='%s' instance='%s/%s'",
		prefix, e.ID, e.Type, e.Instance.Deployment, e.Instance.Group, e.Instance.ID)
}
