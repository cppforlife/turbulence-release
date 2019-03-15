package tasks

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type BlockDNSOptions struct {
	Type    string
	Timeout string // Times may be suffixed with ms,s,m,h
}

func (BlockDNSOptions) _private() {}

type BlockDNSTask struct {
	cmdRunner boshsys.CmdRunner
	opts      BlockDNSOptions
}

func NewBlockDNSTask(
	cmdRunner boshsys.CmdRunner,
	opts BlockDNSOptions,
	_ boshlog.Logger,
) BlockDNSTask {
	return BlockDNSTask{cmdRunner, opts}
}

func (t BlockDNSTask) Execute(stopCh chan struct{}) error {
	timeoutCh, err := NewOptionalTimeoutCh(t.opts.Timeout)
	if err != nil {
		return err
	}

	rules := t.rules()

	for _, r := range rules {
		err := t.iptables("-A", r)
		if err != nil {
			return err
		}
	}

	select {
	case <-timeoutCh:
	case <-stopCh:
	}

	for _, r := range rules {
		err := t.iptables("-D", r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t BlockDNSTask) rules() []string {
	return []string{ "OUTPUT -p tcp --destination-port 53 -j DROP", "OUTPUT -p udp --destination-port 53 -j DROP" }
}

func (t BlockDNSTask) iptables(action, rule string) error {
	args := append([]string{action}, strings.Split(rule, " ")...)

	_, _, _, err := t.cmdRunner.RunCommand("iptables", args...)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to iptables")
	}

	return nil
}
