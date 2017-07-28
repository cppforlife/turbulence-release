package incident

import (
	"encoding/json"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/turbulence/director"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/incident/selector"
	"github.com/cppforlife/turbulence/tasks"
)

type Incident struct {
	director   director.Director
	reporter   reporter.Reporter
	tasksRepo  tasks.Repo
	updateFunc func(Incident) error

	id string

	Tasks    tasks.OptionsSlice
	Selector selector.Req

	executionStartedAt   time.Time
	executionCompletedAt time.Time

	events *reporter.Events

	logTag string
	logger boshlog.Logger
}

func (i Incident) ID() string               { return i.id }
func (i Incident) Events() *reporter.Events { return i.events }

func (i Incident) ExecutionStartedAt() time.Time   { return i.executionStartedAt }
func (i Incident) ExecutionCompletedAt() time.Time { return i.executionCompletedAt }

func (i Incident) HasKillTask() bool {
	for _, task := range i.Tasks {
		if _, ok := task.(tasks.KillOptions); ok {
			return true
		}
	}
	return false
}

func (i Incident) TaskTypes() []string {
	var types []string

	for _, taskOpts := range i.Tasks {
		types = append(types, tasks.OptionsType(taskOpts))
	}

	return types
}

func (i Incident) ShortDescription() (string, error) {
	b, err := json.Marshal(IncidentReq{i.Tasks, i.Selector})
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (i Incident) Description() (string, error) {
	b, err := json.MarshalIndent(IncidentReq{i.Tasks, i.Selector}, "", "    ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
