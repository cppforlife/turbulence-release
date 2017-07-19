package agentreqs

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FirewallOptions struct {
	Type    string
	Timeout string // Times may be suffixed with ms,s,m,h

	BlockBOSHAgent bool
}

type FirewallTask struct {
	cmdRunner boshsys.CmdRunner
	opts      FirewallOptions

	allowedOutputDest []FirewallTaskDest
	logger            boshlog.Logger
}

type FirewallTaskDest struct {
	Host string
	Port int

	IsBOSHMbus bool
}

func NewFirewallTask(
	cmdRunner boshsys.CmdRunner,
	opts FirewallOptions,
	allowedOutputDest []FirewallTaskDest,
	logger boshlog.Logger,
) FirewallTask {
	return FirewallTask{cmdRunner, opts, allowedOutputDest, logger}
}

func (t FirewallTask) Execute() error {
	t.logger.Info("FirewallTask", "not implemented on Windows")
	return nil
}
