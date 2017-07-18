package agentreqs

import (
	"fmt"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FirewallOptions struct {
	Type    string
	Timeout string // Times may be suffixed with ms,s,m,h

	BlockBOSHAgent bool
}

func (FirewallOptions) _private() {}

type FirewallTask struct {
	cmdRunner boshsys.CmdRunner
	opts      FirewallOptions

	allowedOutputDest []FirewallTaskDest
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
	_ boshlog.Logger,
) FirewallTask {
	return FirewallTask{cmdRunner, opts, allowedOutputDest}
}

func (t FirewallTask) Execute() error {
	timeout, err := time.ParseDuration(t.opts.Timeout)
	if err != nil {
		return bosherr.WrapError(err, "Parsing timeout")
	}

	rules := t.rules()

	for _, r := range rules {
		err := t.iptables("-A", r)
		if err != nil {
			return err
		}
	}

	<-time.After(timeout)

	for _, r := range rules {
		err := t.iptables("-D", r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t FirewallTask) rules() []string {
	var rules []string

	// Allow response traffic from allowed destinations
	inputRuleTpl := "INPUT ! -i lo -p tcp -s %s --sport %d -m state --state NEW,ESTABLISHED -j ACCEPT"

	for _, dest := range t.allowedOutputDest {
		if !t.opts.BlockBOSHAgent || !dest.IsBOSHMbus {
			rules = append(rules, fmt.Sprintf(inputRuleTpl, dest.Host, dest.Port))
		}
	}

	// Allow all localhost input traffic; allow SSH traffic; drop rest
	inputRules := []string{
		"INPUT ! -i lo -p tcp --dport 22 -m state --state NEW,ESTABLISHED -j ACCEPT",
		"INPUT ! -i lo -j DROP",
	}

	rules = append(rules, inputRules...)

	// Allow outgoing traffic to allowed destinations
	outputRuleTpl := "OUTPUT ! -o lo -p tcp -d %s --dport %d -m state --state NEW,ESTABLISHED -j ACCEPT"

	for _, dest := range t.allowedOutputDest {
		if !t.opts.BlockBOSHAgent || !dest.IsBOSHMbus {
			rules = append(rules, fmt.Sprintf(outputRuleTpl, dest.Host, dest.Port))
		}
	}

	// Allow all localhost output traffic; allow SSH traffic; drop rest
	outputRules := []string{
		"OUTPUT ! -o lo -p tcp --sport 22 -m state --state NEW,ESTABLISHED -j ACCEPT",
		"OUTPUT ! -o lo -j DROP",
	}

	return append(rules, outputRules...)
}

func (t FirewallTask) iptables(action, rule string) error {
	args := append([]string{action}, strings.Split(rule, " ")...)

	_, _, _, err := t.cmdRunner.RunCommand("iptables", args...)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to iptables")
	}

	return nil
}
