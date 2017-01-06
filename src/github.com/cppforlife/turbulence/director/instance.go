package director

import (
	"fmt"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type InstanceImpl struct {
	deployment boshdir.Deployment

	deploymentName, az, group, id string

	cid, agentID string
}

func (i InstanceImpl) Deployment() string { return i.deployment.Name() }
func (i InstanceImpl) AZ() string         { return i.az }
func (i InstanceImpl) Group() string      { return i.group }
func (i InstanceImpl) ID() string         { return i.id }
func (i InstanceImpl) AgentID() string    { return i.agentID }
func (i InstanceImpl) HasVM() bool        { return len(i.cid) > 0 }

func (i InstanceImpl) DeleteVM() error {
	if !i.HasVM() {
		return fmt.Errorf("Cannot delete VM for instance '%s' since it does not have an associated VM", i.id)
	}

	return i.deployment.DeleteVM(i.cid)
}
