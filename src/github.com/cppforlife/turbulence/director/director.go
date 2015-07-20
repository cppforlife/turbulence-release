package director

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/cloud"
)

type Director struct {
	cpi    cloud.Cloud
	client Client
}

type Deployment struct {
	cpi    cloud.Cloud
	client Client

	Name string
}

type Job struct {
	Name             string
	instancesWithVMs []Instance
}

type Instance struct {
	cpi cloud.Cloud

	Index   int // particular instance index e.g. 0,1,2,3...
	vmCID   string
	AgentID string
}

func (d Director) FindDeployment(name string) (Deployment, error) {
	// An unecessary call but useful to make sure deployment exists
	_, err := d.client.Deployment(name)
	if err != nil {
		return Deployment{}, bosherr.WrapErrorf(err, "Finding deployment '%s'", name)
	}

	return Deployment{Name: name, client: d.client, cpi: d.cpi}, nil
}

func (d Deployment) FindJobs(name string) ([]Job, error) {
	jobs, err := d.Jobs()
	if err != nil {
		return nil, err
	}

	var matchedJobs []Job

	for _, job := range jobs {
		matched, err := filepath.Match(name, job.Name)
		if err != nil {
			return nil, err
		}

		if matched {
			matchedJobs = append(matchedJobs, job)
		}
	}

	if len(matchedJobs) == 0 {
		return nil, bosherr.Errorf("Job '%s' does match any jobs.", name)
	}

	return matchedJobs, nil
}

func (d Deployment) Jobs() ([]Job, error) {
	var jobs []Job

	vms, err := d.client.DeploymentVMs(d.Name)
	if err != nil {
		return jobs, bosherr.WrapErrorf(err, "Listing VMs in a deployment '%s'", d.Name)
	}

	for _, vm := range vms {
		foundJobIndex := -1
		job := Job{Name: vm.JobName}

		for jobI, j := range jobs {
			if j.Name == vm.JobName {
				foundJobIndex = jobI
				job = jobs[jobI]
				break
			}
		}

		instance := Instance{
			cpi: d.cpi,

			Index:   vm.JobIndex,
			vmCID:   vm.CID,
			AgentID: vm.AgentID,
		}

		job.instancesWithVMs = append(job.instancesWithVMs, instance)

		if foundJobIndex >= 0 {
			jobs[foundJobIndex] = job
		} else {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func (j Job) InstancesWithVMs() ([]Instance, error) {
	return j.instancesWithVMs, nil
}

func (i Instance) DeleteVM() error {
	return i.cpi.DeleteVM(i.vmCID)
}
