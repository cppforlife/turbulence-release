package client

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/incident"
)

type TurbulenceImpl struct {
	client Client
}

func (t TurbulenceImpl) CreateIncident(req incident.Request) (Incident, error) {
	resp, err := t.client.CreateIncident(req)
	if err != nil {
		return nil, err
	}

	incident := &IncidentImpl{
		client: t.client,
		id:     resp.ID,
		resp:   resp,
	}

	return incident, nil
}

func (c Client) CreateIncident(req incident.Request) (incident.Response, error) {
	var resp incident.Response

	err := c.clientRequest.Post("/api/v1/incidents", req, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Creating incident")
	}

	return resp, nil
}
