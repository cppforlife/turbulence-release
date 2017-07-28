package scheduledinc

import (
	"encoding/json"
	"fmt"

	"github.com/cppforlife/turbulence/incident"
)

type Request struct {
	Schedule string
	Incident incident.Request
}

type Response struct {
	ID string

	Schedule string
	Incident incident.Request
}

type Responses []Response

func NewResponses(sis []ScheduledIncident) Responses {
	resp := []Response{}

	for _, si := range sis {
		resp = append(resp, NewResponse(si))
	}

	return resp
}

func NewResponse(si ScheduledIncident) Response {
	return Response{
		ID: si.ID,

		Schedule: si.Schedule,
		Incident: si.Incident,
	}
}

func (r Response) URL() string {
	return fmt.Sprintf("/scheduled_incidents/%s", r.ID)
}

func (r Response) Description() (string, error) {
	b, err := json.MarshalIndent(r.Incident, "", "    ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
