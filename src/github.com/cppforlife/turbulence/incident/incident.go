package incident

import (
	"encoding/json"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/turbulence/agentreqs"
	"github.com/cppforlife/turbulence/director"
)

type Incident struct {
	director      director.Director
	reporter      Reporter
	agentReqsRepo agentreqs.Repo
	updateFunc    func(Incident) error

	ID string

	Tasks       agentreqs.TaskOptionsSlice
	Deployments []Deployment

	ExecutionStartedAt   time.Time
	ExecutionCompletedAt time.Time

	Events *Events

	logTag string
	logger boshlog.Logger
}

func (i Incident) Execute() error {
	i.logger.Debug(i.logTag, "Executing incident '%s'", i.ID)

	i.ExecutionStartedAt = time.Now().UTC()

	err := i.updateFunc(i)
	if err != nil {
		return bosherr.Errorf("Updating execution started at")
	}

	i.reporter.ReportIncidentExecutionStart(i)
	i.executeOnDeployments()

	i.logger.Debug(i.logTag, "Waiting for incident '%s' events completion", i.ID)

	// Serialize updates to the incident and events
	for r := range i.Events.Results() {
		r.Event.MarkError(r.Error)
		i.update()
	}

	i.logger.Debug(i.logTag, "Incident '%s' events completed", i.ID)

	i.ExecutionCompletedAt = time.Now().UTC()

	i.update()
	i.reporter.ReportIncidentExecutionCompletion(i)

	i.logger.Debug(i.logTag, "Incident '%s' completed", i.ID)

	return i.Events.FirstError()
}

func (i Incident) executeOnDeployments() {
	for _, depl := range i.Deployments {
		event := i.Events.Add(Event{
			Type:           EventTypeFindDeployment,
			DeploymentName: depl.Name,
		})
		actualDeployment, err := i.director.FindDeployment(depl.Name)
		if event.MarkError(err) {
			continue
		}

		for _, job := range depl.Jobs {
			event = i.Events.Add(Event{
				Type:           EventTypeFindJobs,
				DeploymentName: depl.Name,
				JobNameMatch:   job.Name,
			})
			actualJobs, err := actualDeployment.FindJobs(job.Name)
			if event.MarkError(err) {
				continue
			}

			for _, actualJob := range actualJobs {
				event = i.Events.Add(Event{
					Type:           EventTypeFindInstances,
					DeploymentName: depl.Name,
					JobName:        actualJob.Name,
				})
				actualInstances, err := actualJob.InstancesWithVMs()
				if event.MarkError(err) {
					continue
				}

				event = i.Events.Add(Event{
					Type:           EventTypeSelectInstances,
					DeploymentName: depl.Name,
					JobName:        actualJob.Name,
				})
				selectedIndices, err := job.SelectedIndices(len(actualInstances))
				if event.MarkError(err) {
					continue
				}

				i.logger.Debug(i.logTag, "Selected indices '%v' for job '%s'", selectedIndices, actualJob.Name)

				for _, index := range selectedIndices {
					actualInstance := actualInstances[index]

					eventTpl := Event{
						DeploymentName: depl.Name,
						JobName:        actualJob.Name,
						JobIndex:       &actualInstance.Index,
					}

					// Ignore all other tasks if we are planning to kill the VM
					if i.HasKillTask() {
						i.killInstance(eventTpl, actualInstance)
					} else {
						i.executeTasks(eventTpl, actualInstance.AgentID)
					}
				}

				// Add events for all instances
				i.update()
			}
		}
	}
}

func (i Incident) killInstance(eventTpl Event, actualInstance director.Instance) {
	eventTpl.Type = agentreqs.TaskOptsType(agentreqs.KillOptions{})

	event := i.Events.Add(eventTpl)

	go func() {
		err := actualInstance.DeleteVM()
		i.Events.RegisterResult(EventResult{event, err})
	}()
}

func (i Incident) executeTasks(eventTpl Event, agentID string) {
	var tasks []agentreqs.Task
	var events []*Event

	for _, taskOpts := range i.Tasks {
		eventTpl.Type = agentreqs.TaskOptsType(taskOpts)

		event := i.Events.Add(eventTpl)

		if len(event.ID) == 0 {
			event.MarkError(bosherr.Error("Empty event ID cannot be used for as an agent task ID"))
			continue
		}

		task := agentreqs.Task{
			ID:       event.ID,
			Optionss: []agentreqs.TaskOptions{taskOpts},
		}

		tasks = append(tasks, task)
		events = append(events, event)
	}

	go func() {
		err := i.agentReqsRepo.QueueAndWait(agentID, tasks)
		if err != nil {
			i.logger.Error(i.logTag, "Failed to queue/wait for agent '%s': %s", agentID, err.Error())

			for _, event := range events {
				i.Events.RegisterResult(EventResult{event, err})
			}

			return
		}

		for _, event := range events {
			go func() {
				_, err := i.agentReqsRepo.Wait(event.ID)
				i.Events.RegisterResult(EventResult{event, err})
			}()
		}
	}()
}

func (i Incident) update() {
	err := i.updateFunc(i)
	if err != nil {
		i.logger.Error(i.logTag, "Failed to update incident '%s': %s", i.ID, err.Error())
	}
}

func (i Incident) HasKillTask() bool {
	for _, task := range i.Tasks {
		if _, ok := task.(agentreqs.KillOptions); ok {
			return true
		}
	}

	return false
}

func (i Incident) TaskTypes() []string {
	var types []string

	for _, taskOpts := range i.Tasks {
		types = append(types, agentreqs.TaskOptsType(taskOpts))
	}

	return types
}

func (i Incident) ShortDescription() (string, error) {
	b, err := json.Marshal(IncidentReq{i.Tasks, i.Deployments})
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (i Incident) Description() (string, error) {
	b, err := json.MarshalIndent(IncidentReq{i.Tasks, i.Deployments}, "", "    ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
