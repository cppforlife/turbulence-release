package scheduledinc

import (
	"encoding/json"
	"fmt"

	"github.com/cppforlife/turbulence/incident"
)

type ScheduledRequest struct {
	Schedule string
	Incident incident.Request
}

type ScheduledIncidentResp struct {
	ID string

	Schedule string
	Incident incident.Request
}

type ScheduledIncidentsResp []ScheduledIncidentResp

func NewScheduledIncidentsResp(sis []ScheduledIncident) ScheduledIncidentsResp {
	resp := []ScheduledIncidentResp{}

	for _, si := range sis {
		resp = append(resp, NewScheduledIncidentResp(si))
	}

	return resp
}

func NewScheduledIncidentResp(si ScheduledIncident) ScheduledIncidentResp {
	return ScheduledIncidentResp{
		ID: si.ID,

		Schedule: si.Schedule,
		Incident: si.Incident,
	}
}

func (r ScheduledIncidentResp) URL() string {
	return fmt.Sprintf("/scheduled_incidents/%s", r.ID)
}

func (r ScheduledIncidentResp) Description() (string, error) {
	b, err := json.MarshalIndent(r.Incident, "", "    ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
