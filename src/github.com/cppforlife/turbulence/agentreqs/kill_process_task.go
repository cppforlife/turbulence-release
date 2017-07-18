package agentreqs

import (
	"math/rand"
	"path/filepath"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cppforlife/turbulence/agentreqs/monit"
)

type KillProcessOptions struct {
	Type string

	// Optionally specify any process pattern used with pkill;
	// takes precedence over monitored processes
	ProcessName string

	// Optionally specify monitored process name
	MonitoredProcessName string

	// If names are empty, randomly selected monitored process is killed
}

func (KillProcessOptions) _private() {}

type KillProcessTask struct {
	monitClient monit.Client
	cmdRunner   boshsys.CmdRunner
	opts        KillProcessOptions

	logTag string
	logger boshlog.Logger
}

func NewKillProcessTask(
	monitClient monit.Client,
	cmdRunner boshsys.CmdRunner,
	opts KillProcessOptions,
	logger boshlog.Logger,
) KillProcessTask {
	return KillProcessTask{monitClient, cmdRunner, opts, "agentreqs.KillProcessTask", logger}
}

func (t KillProcessTask) Execute() error {
	if len(t.opts.ProcessName) > 0 {
		return t.killProcesses(t.opts.ProcessName)
	}

	if len(t.opts.MonitoredProcessName) > 0 {
		return t.killMatchingServices(t.opts.MonitoredProcessName)
	}

	return t.killRandomService()
}

func (t KillProcessTask) killProcesses(name string) error {
	t.logger.Debug(t.logTag, "Killing processes matching '%s'", name)

	_, _, _, err := t.cmdRunner.RunCommand("pkill", "-9", name)
	if err != nil {
		return bosherr.WrapError(err, "Killing processes")
	}

	return nil
}

func (t KillProcessTask) killMatchingServices(name string) error {
	services, err := t.monitClient.Services()
	if err != nil {
		return bosherr.WrapError(err, "Getting monit services")
	}

	var matchedServices []monit.Service

	for _, service := range services {
		matched, err := filepath.Match(name, service.Name)
		if err != nil {
			return err
		}

		if matched {
			matchedServices = append(matchedServices, service)
		}
	}

	if len(matchedServices) == 0 {
		return bosherr.Errorf("Process '%s' must match at least one monitored process", name)
	}

	var firstErr error

	for _, service := range matchedServices {
		err := t.killService(service)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (t KillProcessTask) killRandomService() error {
	services, err := t.monitClient.Services()
	if err != nil {
		return bosherr.WrapError(err, "Getting monit services")
	}

	if len(services) == 0 {
		return bosherr.Error("At least one monitored process must be present")
	}

	return t.killService(services[rand.Intn(len(services))])
}

func (t KillProcessTask) killService(service monit.Service) error {
	t.logger.Debug(t.logTag, "Killing process '%s' (PID: %d)", service.Name, service.PID)

	if service.PID == 0 {
		return bosherr.Errorf("Process '%s' PID was 0 which is not a valid PID", service.Name)
	}

	if service.PID == 1 {
		return bosherr.Errorf("Process '%s' PID was 1 which is not allowed to kill", service.Name)
	}

	_, _, _, err := t.cmdRunner.RunCommand("kill", "-9", strconv.Itoa(service.PID))
	if err != nil {
		return bosherr.WrapError(err, "Killing process")
	}

	return nil
}
