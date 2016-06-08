package agentreqs

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FillDiskOptions struct {
	Type string

	// By default root disk will be filled
	Persistent bool
	Ephemeral  bool
	Temporary  bool
}

type FillDiskTask struct {
	cmdRunner boshsys.CmdRunner
	opts      FillDiskOptions
}

func NewFillDiskTask(cmdRunner boshsys.CmdRunner, opts FillDiskOptions, _ boshlog.Logger) FillDiskTask {
	return FillDiskTask{cmdRunner, opts}
}

func (t FillDiskTask) Execute() error {
	if t.opts.Persistent {
		return t.fill("/var/vcap/store/.filler")
	}

	if t.opts.Ephemeral {
		return t.fill("/var/vcap/data/.filler")
	}

	if t.opts.Temporary {
		return t.fill("/tmp/.filler")
	}

	return t.fill("/.filler")
}

func (t FillDiskTask) fill(path string) error {
	_, _, _, err := t.cmdRunner.RunCommand("dd", "if=/dev/zero", "of="+path, "bs=1M")
	if err != nil {
		return bosherr.WrapError(err, "Filling disk")
	}

	return nil
}
