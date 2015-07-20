package director

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Client struct {
	clientRequest clientRequest
}

type DeploymentResp struct {
	Manifest string
}

type VMResp struct {
	JobName  string `json:"job"`   // e.g. dummy1
	JobIndex int    `json:"index"` // e.g. 0,1,2

	AgentID string `json:"agent_id"` // e.g. 3b30123e-dfa6-4eff-abe6-63c2d5a88938
	CID     string // e.g. vm-ce10ae6a-6c31-413b-a134-7179f49e0bda
}

func NewClient(endpoint string, httpClient boshhttp.HTTPClient, logger boshlog.Logger) Client {
	clientRequest := clientRequest{
		endpoint:   endpoint,
		httpClient: httpClient,
		logger:     logger,
	}

	return Client{clientRequest: clientRequest}
}

func (c Client) Deployment(deploymentName string) (DeploymentResp, error) {
	var dep DeploymentResp

	path := fmt.Sprintf("/deployments/%s", deploymentName)

	err := c.clientRequest.Get(path, &dep)
	if err != nil {
		return dep, bosherr.WrapErrorf(err, "Finding deployment '%s'", deploymentName)
	}

	return dep, nil
}

func (c Client) DeploymentVMs(deploymentName string) ([]VMResp, error) {
	var vms []VMResp

	path := fmt.Sprintf("/deployments/%s/vms", deploymentName)

	err := c.clientRequest.Get(path, &vms)
	if err != nil {
		return vms, bosherr.WrapErrorf(err, "Listing deployment '%s' VMs", deploymentName)
	}

	return vms, nil
}
