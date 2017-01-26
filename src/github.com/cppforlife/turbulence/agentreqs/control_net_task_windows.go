package agentreqs

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

// TODO (CEV): Refactor so that we don't duplicate the unix type definition.
//
// See http://www.linuxfoundation.org/collaborate/workgroups/networking/netem
type ControlNetOptions struct {
	Type    string
	Timeout string // Times may be suffixed with ms,s,m,h

	// slow: tc qdisc add dev eth0 root netem delay 50ms 10ms distribution normal
	Delay          string
	DelayVariation string

	// flaky: tc qdisc add dev eth0 root netem loss 20% 75%
	Loss            string
	LossCorrelation string

	// reset: tc qdisc del dev eth0 root
}

type ControlNetTask struct {
	cmdRunner boshsys.CmdRunner
	opts      ControlNetOptions
	logger    boshlog.Logger
}

func NewControlNetTask(cmdRunner boshsys.CmdRunner, opts ControlNetOptions, logger boshlog.Logger) ControlNetTask {
	return ControlNetTask{cmdRunner: cmdRunner, opts: opts, logger: logger}
}

func (t ControlNetTask) Execute() error {
	t.logger.Info("ControlNetTask", "Not implemented on Windows")
	return nil
}
