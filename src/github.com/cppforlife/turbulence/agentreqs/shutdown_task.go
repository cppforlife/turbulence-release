package agentreqs

import (
	"fmt"
	"runtime"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type ShutdownOptions struct {
	Type string

	// By default system will be powered off
	Reboot bool
	Force  bool

	Crash bool
	Sysrq string
}

type ShutdownTask struct {
	cmdRunner boshsys.CmdRunner
	opts      ShutdownOptions
	logger    boshlog.Logger
}

func NewShutdownTask(cmdRunner boshsys.CmdRunner, opts ShutdownOptions, logger boshlog.Logger) ShutdownTask {
	return ShutdownTask{cmdRunner, opts, logger}
}

func (t ShutdownTask) Execute() error {
	if t.opts.Crash {
		return t.crash()
	}

	if len(t.opts.Sysrq) > 0 {
		return t.sysrq(t.opts.Sysrq)
	}

	if t.opts.Reboot {
		return t.reboot(t.opts.Force)
	}

	return t.halt(t.opts.Force)
}

func (t ShutdownTask) crash() error {
	if runtime.GOOS == "windows" {
		t.logger.Info("ShutdownTask", "crash not implemented on Windows - stopping with force")
		return t.halt(true)
	}
	return t.sysrq("c")
}

func (t ShutdownTask) sysrq(val string) error {
	if runtime.GOOS == "windows" {
		t.logger.Info("ShutdownTask", "syrq not implemented on Windows")
		return nil
	}

	cmd := fmt.Sprintf("echo 1 > /proc/sys/kernel/sysrq && echo %s > /proc/sysrq-trigger", val)

	_, _, _, err := t.cmdRunner.RunCommand("bash", "-c", cmd)
	if err != nil {
		return bosherr.WrapError(err, "Sysrq trigger")
	}

	return nil
}

func (t ShutdownTask) halt(force bool) error {
	var args []string
	var command string
	if runtime.GOOS == "windows" {
		command = "-Command"
		args = []string{"Stop-Computer"}
		if force {
			args = append(args, "-Force")
		}
	} else {
		command = "halt"
		if force {
			args = append(args, "--force")
		}
	}

	_, _, _, err := t.cmdRunner.RunCommand(command, args...)
	if err != nil {
		return bosherr.WrapError(err, "Halting")
	}

	return nil
}

func (t ShutdownTask) reboot(force bool) error {
	var args []string
	var command string
	if runtime.GOOS == "windows" {
		command = "-Command"
		args = []string{"Restart-Computer"}
		if force {
			args = append(args, "-Force")
		}
	} else {
		command = "reboot"
		if force {
			args = append(args, "--force")
		}
	}

	_, _, _, err := t.cmdRunner.RunCommand(command, args...)
	if err != nil {
		return bosherr.WrapError(err, "Rebooting")
	}

	return nil
}
