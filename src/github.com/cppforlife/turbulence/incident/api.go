package incident

import (
	"fmt"
	"strings"
	"time"

	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/incident/selector"
	"github.com/cppforlife/turbulence/tasks"
)

type IncidentReq struct {
	Tasks    tasks.OptionsSlice
	Selector selector.Req
}

type IncidentResp struct {
	incident Incident

	ID string

	Tasks    tasks.OptionsSlice
	Selector selector.Req

	ExecutionStartedAt   string
	ExecutionCompletedAt string

	Events []reporter.EventResp

	description string
}

type IncidentsResp []IncidentResp

func NewIncidentsResp(incidents []Incident) IncidentsResp {
	resp := []IncidentResp{}

	for _, incid := range incidents {
		resp = append(resp, NewIncidentResp(incid))
	}

	return resp
}

func NewIncidentResp(incident Incident) IncidentResp {
	var eventResps []reporter.EventResp

	for _, event := range incident.Events().Events() {
		eventResps = append(eventResps, reporter.NewEventResp(event))
	}

	var completedAt string

	if (incident.ExecutionCompletedAt() != time.Time{}) {
		completedAt = incident.ExecutionCompletedAt().Format(time.RFC3339)
	}

	return IncidentResp{
		incident: incident,

		ID: incident.ID(),

		Tasks:    incident.Tasks,
		Selector: incident.Selector,

		ExecutionStartedAt:   incident.ExecutionStartedAt().Format(time.RFC3339),
		ExecutionCompletedAt: completedAt,

		Events: eventResps,
	}
}

func (r IncidentResp) URL() string { return fmt.Sprintf("/incidents/%s", r.ID) }

func (r IncidentResp) TaskTypes() string { return strings.Join(r.incident.TaskTypes(), ", ") }

func (r IncidentResp) Description() (string, error) { return r.incident.Description() }

func (r IncidentResp) HasEventErrors() bool {
	for _, eventResp := range r.Events {
		if len(eventResp.Error) > 0 {
			return true
		}
	}

	return false
}
