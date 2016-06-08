package main

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cppforlife/turbulence/agentreqs"
	"github.com/cppforlife/turbulence/agentreqs/monit"
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
	Execute() error
}

type AgentConfig struct {
	APIHost string
	APIPort int

	BOSHMbusHost string
	BOSHMbusPort int
}

func (c AgentConfig) AllowedOutputDests() []agentreqs.FirewallTaskDest {
	return []agentreqs.FirewallTaskDest{
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

func (a Agent) executeTask(task agentreqs.Task) {
	a.logger.Debug(a.logTag, "Received agent task options '%#v'", task)

	var t agentTask
	var err error

	switch opts := task.Optionss[0].(type) {
	case agentreqs.KillProcessOptions:
		monitClient, err := a.monitProvider.Get()
		if err != nil {
			err = bosherr.WrapError(err, "Failed to retrieve monit client")
		} else {
			t = agentreqs.NewKillProcessTask(monitClient, a.cmdRunner, opts, a.logger)
		}

	case agentreqs.StressOptions:
		t = agentreqs.NewStressTask(a.cmdRunner, opts, a.logger)

	case agentreqs.ControlNetOptions:
		t = agentreqs.NewControlNetTask(a.cmdRunner, opts, a.logger)

	case agentreqs.FirewallOptions:
		t = agentreqs.NewFirewallTask(a.cmdRunner, opts, a.agentConfig.AllowedOutputDests(), a.logger)

	case agentreqs.FillDiskOptions:
		t = agentreqs.NewFillDiskTask(a.cmdRunner, opts, a.logger)

	case agentreqs.ShutdownOptions:
		t = agentreqs.NewShutdownTask(a.cmdRunner, opts, a.logger)

	default:
		err = bosherr.Errorf("Unknown agent task '%T'", task.Optionss[0])
		a.logger.Error(a.logTag, "Ignoring unknown agent task '%T'", task.Optionss[0])
	}

	if t != nil {
		err = t.Execute()
		if err != nil {
			a.logger.Error(a.logTag, "Failed executing agent task: %s", err.Error())
		}
	}

	err = a.client.UpdateTask(task.ID, err)
	if err != nil {
		a.logger.Error(a.logTag, "Failed updating agent task: %s", err.Error())
	}
}
