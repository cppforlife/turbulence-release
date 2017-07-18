package agentreqs

import (
	"fmt"

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

func (ShutdownOptions) _private() {}

type ShutdownTask struct {
	cmdRunner boshsys.CmdRunner
	opts      ShutdownOptions
}

func NewShutdownTask(cmdRunner boshsys.CmdRunner, opts ShutdownOptions, _ boshlog.Logger) ShutdownTask {
	return ShutdownTask{cmdRunner, opts}
}

func (t ShutdownTask) Execute() error {
	if t.opts.Crash {
		return t.sysrq("c")
	}

	if len(t.opts.Sysrq) > 0 {
		return t.sysrq(t.opts.Sysrq)
	}

	if t.opts.Reboot {
		return t.reboot(t.opts.Force)
	}

	return t.halt(t.opts.Force)
}

func (t ShutdownTask) sysrq(val string) error {
	cmd := fmt.Sprintf("echo 1 > /proc/sys/kernel/sysrq && echo %s > /proc/sysrq-trigger", val)

	_, _, _, err := t.cmdRunner.RunCommand("bash", "-c", cmd)
	if err != nil {
		return bosherr.WrapError(err, "Sysrq trigger")
	}

	return nil
}

func (t ShutdownTask) halt(force bool) error {
	var forceArg []string

	if force {
		forceArg = append(forceArg, "--force")
	}

	_, _, _, err := t.cmdRunner.RunCommand("halt", forceArg...)
	if err != nil {
		return bosherr.WrapError(err, "Halting")
	}

	return nil
}

func (t ShutdownTask) reboot(force bool) error {
	var forceArg []string

	if force {
		forceArg = append(forceArg, "--force")
	}

	_, _, _, err := t.cmdRunner.RunCommand("reboot", forceArg...)
	if err != nil {
		return bosherr.WrapError(err, "Rebooting")
	}

	return nil
}
