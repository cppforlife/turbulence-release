package main

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cppforlife/turbulence/tasks"
	"github.com/cppforlife/turbulence/tasks/monit"
)

type Agent struct {
	agentID     string
	agentConfig AgentConfig

	client        Client
	monitProvider monit.ClientProvider
	cmdRunner     boshsys.CmdRunner

	logTag string
	logger boshlog.Logger
}

type agentTask interface {
	Execute(stopCh chan struct{}) error
}

type AgentConfig struct {
	APIHost string
	APIPort int

	BOSHMbusHost string
	BOSHMbusPort int
}

func (c AgentConfig) AllowedOutputDests() []tasks.FirewallTaskDest {
	return []tasks.FirewallTaskDest{
		{Host: c.APIHost, Port: c.APIPort},
		{Host: c.BOSHMbusHost, Port: c.BOSHMbusPort, IsBOSHMbus: true},
	}
}

func newAgent(
	agentID string,
	agentConfig AgentConfig,
	client Client,
	monitProvider monit.ClientProvider,
	cmdRunner boshsys.CmdRunner,
	logger boshlog.Logger,
) Agent {
	return Agent{
		agentID:     agentID,
		agentConfig: agentConfig,

		client:        client,
		monitProvider: monitProvider,
		cmdRunner:     cmdRunner,

		logTag: "Agent",
		logger: logger,
	}
}

func (a Agent) ContiniouslyExecuteTasks() error {
	a.logger.Info(a.logTag, "Started continiously executing tasks")

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			tasks, err := a.client.FetchTasks(a.agentID)
			if err != nil {
				a.logger.Error(a.logTag, "Failed fetching tasks: %s", err.Error())
			} else {
				// Execute tasks in parallel
				for _, task := range tasks {
					go a.executeTask(task)
				}
			}
		}
	}
}

func (a Agent) executeTask(task tasks.Task) {
	a.logger.Debug(a.logTag, "Received agent task options '%#v'", task)

	task1, err := a.buildAgentTask(task)

	if task1 != nil && err == nil {
		stopCh := make(chan struct{}, 1) // allow one stop
		endPollCh := make(chan struct{}, 1)

		go func() {
			for {
				select {
				case <-endPollCh:
					return
				default:
				}

				resp, err := a.client.FetchTaskState(task.ID)
				if err != nil {
					a.logger.Error(a.logTag, "Failed fetching agent task control: %s", err.Error())
				}

				if resp.Stop {
					close(stopCh)
					return
				}

				time.Sleep(1 * time.Second)
			}
		}()

		err = task1.Execute(stopCh)
		if err != nil {
			err = bosherr.WrapError(err, "Task execution")
			a.logger.Error(a.logTag, "Failed executing agent task: %s", err.Error())
		}

		close(endPollCh)
	}

	err = a.client.RecordTaskResult(task.ID, err)
	if err != nil {
		a.logger.Error(a.logTag, "Failed updating agent task: %s", err.Error())
	}
}

func (a Agent) buildAgentTask(task tasks.Task) (agentTask, error) {
	var t agentTask
	var err error

	switch opts := task.Options().(type) {
	case tasks.KillProcessOptions:
		monitClient, err := a.monitProvider.Get()
		if err != nil {
			err = bosherr.WrapError(err, "Failed to retrieve monit client")
		} else {
			t = tasks.NewKillProcessTask(monitClient, a.cmdRunner, opts, a.logger)
		}

	case tasks.StressOptions:
		t = tasks.NewStressTask(a.cmdRunner, opts, a.logger)

	case tasks.ControlNetOptions:
		t = tasks.NewControlNetTask(a.cmdRunner, opts, a.logger)

	case tasks.FirewallOptions:
		t = tasks.NewFirewallTask(a.cmdRunner, opts, a.agentConfig.AllowedOutputDests(), a.logger)

	case tasks.FillDiskOptions:
		t = tasks.NewFillDiskTask(a.cmdRunner, opts, a.logger)

	case tasks.ShutdownOptions:
		t = tasks.NewShutdownTask(a.cmdRunner, opts, a.logger)

	default:
		err = bosherr.Errorf("Unknown agent task '%T'", task.Optionss[0])
		a.logger.Error(a.logTag, "Ignoring unknown agent task '%T'", task.Optionss[0])
	}

	return t, err
}
