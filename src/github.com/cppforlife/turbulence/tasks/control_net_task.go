package tasks

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

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

func (ControlNetOptions) _private() {}

type ControlNetTask struct {
	cmdRunner boshsys.CmdRunner
	opts      ControlNetOptions
}

func NewControlNetTask(cmdRunner boshsys.CmdRunner, opts ControlNetOptions, _ boshlog.Logger) ControlNetTask {
	return ControlNetTask{cmdRunner, opts}
}

func (t ControlNetTask) Execute() error {
	timeout, err := time.ParseDuration(t.opts.Timeout)
	if err != nil {
		return bosherr.WrapError(err, "Parsing timeout")
	}

	if len(t.opts.Delay) == 0 && len(t.opts.Loss) == 0 {
		return bosherr.Error("Must specify delay or loss")
	}

	ifaceNames, err := NonLocalIfaceNames()
	if err != nil {
		return err
	}

	if len(t.opts.Delay) > 0 {
		variation := t.opts.DelayVariation

		if len(variation) == 0 {
			variation = "10ms"
		}

		for _, ifaceName := range ifaceNames {
			err := t.configureDelay(ifaceName, t.opts.Delay, variation)
			if err != nil {
				return err
			}
		}
	}

	if len(t.opts.Loss) > 0 {
		correlation := t.opts.LossCorrelation

		if len(correlation) == 0 {
			correlation = "75%"
		}

		for _, ifaceName := range ifaceNames {
			err := t.configurePacketLoss(ifaceName, t.opts.Loss, correlation)
			if err != nil {
				return err
			}
		}
	}

	<-time.After(timeout)

	for _, ifaceName := range ifaceNames {
		err := t.resetIface(ifaceName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t ControlNetTask) configureDelay(ifaceName, delay, variation string) error {
	args := []string{
		"qdisc", "add", "dev", ifaceName, "root",
		"netem", "delay", delay, variation, "distribution", "normal",
	}

	_, _, _, err := t.cmdRunner.RunCommand("tc", args...)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to tc to add delay")
	}

	return nil
}

func (t ControlNetTask) configurePacketLoss(ifaceName, percent, correlation string) error {
	args := []string{
		"qdisc", "add", "dev", ifaceName, "root",
		"netem", "loss", percent, correlation,
	}

	_, _, _, err := t.cmdRunner.RunCommand("tc", args...)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to tc to add packet loss")
	}

	return nil
}

func (t ControlNetTask) resetIface(ifaceName string) error {
	_, _, _, err := t.cmdRunner.RunCommand("tc", "qdisc", "del", "dev", ifaceName, "root")
	if err != nil {
		return bosherr.WrapError(err, "Resetting tc")
	}

	return nil
}
