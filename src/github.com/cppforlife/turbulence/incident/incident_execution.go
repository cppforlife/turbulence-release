package incident

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/agentreqs"
	"github.com/cppforlife/turbulence/director"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/incident/selector"
)

func (i Incident) Execute() error {
	i.logger.Debug(i.logTag, "Executing incident '%s'", i.id)
	i.executionStartedAt = time.Now().UTC()

	err := i.updateFunc(i)
	if err != nil {
		return bosherr.Errorf("Updating execution started at")
	}

	i.reporter.ReportIncidentExecutionStart(i)
	i.logger.Debug(i.logTag, "Waiting for incident '%s' events completion", i.id)

	i.executeTasks()

	// Serialize updates to the incident and events
	for r := range i.events.Results() {
		r.Event.MarkError(r.Error)
		i.update()
	}

	i.logger.Debug(i.logTag, "Incident '%s' events completed", i.id)

	i.executionCompletedAt = time.Now().UTC()
	i.update()

	i.reporter.ReportIncidentExecutionCompletion(i)
	i.logger.Debug(i.logTag, "Incident '%s' completed", i.id)

	return i.events.FirstError()
}

func (i Incident) executeTasks() {
	event := i.events.Add(reporter.Event{Type: reporter.EventTypeFind})
	instances, err := i.director.AllInstances()
	if event.MarkError(err) {
		return
	}

	var selectedInstances []selector.Instance

	for _, inst := range instances {
		selectedInstances = append(selectedInstances, inst)
	}

	event = i.events.Add(reporter.Event{Type: reporter.EventTypeSelect})
	selectedInstances, err = i.Selector.AsSelector().Select(selectedInstances)
	if event.MarkError(err) {
		return
	}

	for _, inst := range selectedInstances {
		eventTpl := reporter.Event{
			Instance: reporter.EventInstance{
				ID:         inst.ID(),
				Group:      inst.Group(),
				Deployment: inst.Deployment(),
				AZ:         inst.AZ(),
			},
		}
		// Ignore all other tasks if we are planning to kill the VM
		// todo raise error if other tasks are provided?
		if i.HasKillTask() {
			i.killInstance(eventTpl, inst.(director.Instance))
		} else {
			i.executeNonKillTasks(eventTpl, inst.(director.Instance))
		}
	}

	i.update()
}

func (i Incident) executeNonKillTasks(eventTpl reporter.Event, instance director.Instance) {
	var tasks []agentreqs.Task
	var events []*reporter.Event

	for _, taskOpts := range i.Tasks {
		eventTpl.Type = agentreqs.TaskOptsType(taskOpts)

		event := i.events.Add(eventTpl)

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
		err := i.agentReqsRepo.QueueAndWait(instance.AgentID(), tasks)
		if err != nil {
			i.logger.Error(i.logTag, "Failed to queue/wait for agent '%s': %s", instance.AgentID(), err.Error())

			for _, event := range events {
				i.events.RegisterResult(reporter.EventResult{event, err})
			}

			return
		}

		for _, event := range events {
			go func() {
				_, err := i.agentReqsRepo.Wait(event.ID)
				i.events.RegisterResult(reporter.EventResult{event, err})
			}()
		}
	}()
}

func (i Incident) killInstance(eventTpl reporter.Event, instance director.Instance) {
	eventTpl.Type = agentreqs.TaskOptsType(agentreqs.KillOptions{})

	event := i.events.Add(eventTpl)

	go func() {
		err := instance.DeleteVM()
		i.events.RegisterResult(reporter.EventResult{event, err})
	}()
}

func (i Incident) update() {
	err := i.updateFunc(i)
	if err != nil {
		i.logger.Error(i.logTag, "Failed to update incident '%s': %s", i.id, err.Error())
	}
}
