package deployment

import (
	bidisk "github.com/cloudfoundry/bosh-init/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-init/deployment/instance"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Manager interface {
	FindCurrent() (deployment Deployment, found bool, err error)
	Cleanup(biui.Stage) error
}

type manager struct {
	instanceManager   biinstance.Manager
	diskManager       bidisk.Manager
	stemcellManager   bistemcell.Manager
	deploymentFactory Factory
}

func NewManager(
	instanceManager biinstance.Manager,
	diskManager bidisk.Manager,
	stemcellManager bistemcell.Manager,
	deploymentFactory Factory,
) Manager {
	return &manager{
		instanceManager:   instanceManager,
		diskManager:       diskManager,
		stemcellManager:   stemcellManager,
		deploymentFactory: deploymentFactory,
	}
}

func (m *manager) FindCurrent() (deployment Deployment, found bool, err error) {
	instances, err := m.instanceManager.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding current deployment instances")
	}

	disks, err := m.diskManager.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding current deployment disks")
	}

	stemcells, err := m.stemcellManager.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding current deployment stemcells")
	}

	if len(instances) == 0 && len(disks) == 0 && len(stemcells) == 0 {
		return nil, false, nil
	}

	return m.deploymentFactory.NewDeployment(instances, disks, stemcells), true, nil
}

func (m *manager) Cleanup(stage biui.Stage) error {
	if err := m.diskManager.DeleteUnused(stage); err != nil {
		return err
	}

	if err := m.stemcellManager.DeleteUnused(stage); err != nil {
		return err
	}

	return nil
}
