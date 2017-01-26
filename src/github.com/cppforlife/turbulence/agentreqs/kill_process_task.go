package agentreqs

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

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
	if runtime.GOOS == "windows" {
		return t.killProcessesWindows(name)
	}
	return t.killProcessesUnix(name)
}

func (t KillProcessTask) killProcessesUnix(name string) error {
	_, _, _, err := t.cmdRunner.RunCommand("pkill", "-9", name)
	if err != nil {
		return bosherr.WrapError(err, "Killing processes")
	}
	return nil
}

func (t KillProcessTask) listProcessesWindows() (map[string][]int, error) {
	stdout, _, _, err := t.cmdRunner.RunCommand("-Command", "(Get-Process) | foreach {$_.Name, $_.Id}")
	if err != nil {
		return nil, bosherr.WrapError(err, "listing Windows processes")
	}
	procs := make(map[string][]int)
	lines := strings.Split(stdout, "\n")

	for i := 0; i < len(lines)-1; i += 2 {
		name := strings.TrimSpace(lines[i])
		pid, err := strconv.Atoi(strings.TrimSpace(lines[i+1]))
		if name != "" && err == nil && pid > 0 {
			procs[name] = append(procs[name], pid)
		}
	}

	return procs, err
}

func (t KillProcessTask) killProcessesWindows(pattern string) error {
	procs, err := t.listProcessesWindows()
	if err != nil {
		return bosherr.WrapError(err, "KillProcessTask: listing Windows processes")
	}
	t.logger.Info("KillProcessTask", "found %d processes", len(procs))

	pattern = strings.ToLower(pattern)
	foundMatch := false

	for name, pids := range procs {
		// Process names are case-insensitive on Windows
		match, err := filepath.Match(pattern, strings.ToLower(name))
		if err != nil {
			return bosherr.Errorf("Process name pattern '%s' is invalid: %s", name, err)
		}
		if !match {
			continue
		}
		for _, pid := range pids {
			if pid == 0 {
				continue
			}
			if err := killProcess(pid); err != nil {
				t.logger.Info("KillProcessTask", "error killing process: %s - %d", name, pid)
				return bosherr.WrapError(err, "Killing process")
			}
			t.logger.Info("KillProcessTask", "killed process: %s - %d", name, pid)
			foundMatch = true
		}

	}

	if !foundMatch {
		return bosherr.Errorf("Process '%s' must match at least one monitored process", pattern)
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

	if err := killProcess(service.PID); err != nil {
		return bosherr.WrapError(err, "Killing process")
	}

	return nil
}

func killProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	defer proc.Release()
	return proc.Kill()
}
