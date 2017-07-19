package agentreqs

import (
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

func (FillDiskOptions) _private() {}

type FillDiskTask struct {
	cmdRunner boshsys.CmdRunner
	opts      FillDiskOptions
}

func NewFillDiskTask(cmdRunner boshsys.CmdRunner, opts FillDiskOptions, _ boshlog.Logger) FillDiskTask {
	return FillDiskTask{cmdRunner, opts}
}
