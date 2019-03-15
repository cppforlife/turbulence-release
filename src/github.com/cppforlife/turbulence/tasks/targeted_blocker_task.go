// See https://wiki.centos.org/HowTos/Network/IPTables for a good iptables tutorial

package tasks

import (
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type TargetedBlockerOptions struct {
	Type    string
	Timeout string // Times may be suffixed with ms,s,m,h

	Targets []Target
}

var ipPattern = regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(/\d{0,2})?`)
var portPattern = regexp.MustCompile(`\d+(:\d+)?$`)

// Target defines a rule for iptables. Each rule must contain one of {Host, DstPorts, SrcPorts}.
// If DstPorts or SrcPorts ports are included without a DstHost or SrcHost, then those ports will be blocked for all hosts.
// If Host is included without DstPorts or SrcPorts, then all traffic to/from those hosts will be blocked.
type Target struct {
	// Optional destination host to block, can specify an address such as "10.34.4.60", an address block such as "192.168.0.0/24",
	// or a domain name such as "google.com" which will be resolved to an Ip.
	DstHost string

	// Optional source host to block, can specify an address such as "10.34.4.60", an address block such as "192.168.0.0/24",
	// or a domain name such as "google.com" which will be resolved to an Ip.
	SrcHost string

	// Optional direction to block traffic, must be in the set {INPUT, OUTPUT, BOTH}. Defaults to "BOTH".
	Direction string

	// Optional protocol to block, must be in the set {udp, tcp, icmp, all}. Defaults to "all".
	Protocol string

	// Optional "dport" or destination port(s) to block. Specify a single port such as "8080" or a range such as "4530:6740".
	DstPorts string

	// Optional "sport" or source port(s) to block. Specify a single port such as "8080" or a range such as "4530:6740".
	SrcPorts string
}

func (TargetedBlockerOptions) _private() {}

type TargetedBlockerTask struct {
	cmdRunner boshsys.CmdRunner
	opts      TargetedBlockerOptions
	logger    boshlog.Logger
}

func NewTargetedBlockerTask(
	cmdRunner boshsys.CmdRunner,
	opts TargetedBlockerOptions,
	logger boshlog.Logger,
) TargetedBlockerTask {
	return TargetedBlockerTask{cmdRunner, opts, logger}
}

func (t TargetedBlockerTask) Execute(stopCh chan struct{}) error {
	timeoutCh, err := NewOptionalTimeoutCh(t.opts.Timeout)
	if err != nil {
		return err
	}

	rules, err := t.rules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		r := []string{rule[0], "1"} // we want it inserted at the beginning of the rules or it may have no effect.
		r = append(r, rule[1:]...)
		err := t.iptables("-I", r)
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

func (t TargetedBlockerTask) getHost(host string) ([]string, error) {
	if host == "" {
		return nil, nil
	} else if ipPattern.MatchString(host) {
		return ipPattern.FindAllString(host, -1), nil
	} else {
		return t.dig(host)
	}
}

func appendHosts(cmd []string, flag string, hosts ...string) []string {
	ips := ""
	for i, ip := range hosts {
		if i > 0 {
			ips += ","
		}
		ips += ip
	}

	return append(cmd, "-d", ips)
}

func (t TargetedBlockerTask) rules() ([][]string, error) {
	rules := [][]string{}

	for _, target := range t.opts.Targets {
		if target.SrcHost == "" && target.DstHost == "" && target.DstPorts == "" && target.SrcPorts == "" {
			return nil, bosherr.Error("Must specify at least one of SrcHost, DstHost, DstPorts, and or SrcPorts.")
		}

		var dsthosts []string
		var direction, protocol, dports, sports string

		srchosts, err := t.getHost(target.SrcHost)
		if err != nil {
			return nil, err
		}

		dsthosts, err = t.getHost(target.DstHost)
		if err != nil {
			return nil, err
		}

		switch strings.ToUpper(target.Direction) {
		case "INPUT":
			direction = "INPUT"
		case "OUTPUT":
			direction = "OUTPUT"
		case "FORWARD":
			direction = "FORWARD"
		default:
			return nil, bosherr.Errorf("Invalid direction '%v', must be one of {INPUT, OUTPUT, FORWARD}.", target.Direction)
		}

		switch strings.ToLower(target.Protocol) {
		case "":
			protocol = ""
		case "tcp":
			protocol = "tcp"
		case "udp":
			protocol = "udp"
		case "icmp":
			protocol = "icmp"
		case "all":
			protocol = "all"
		default:
			return nil, bosherr.Errorf("Invalid protocol '%v', must be one of {tcp, udp, icmp, all} or blank.", target.Protocol)
		}

		if target.DstPorts == "" {
			dports = ""
		} else if portPattern.MatchString(target.DstPorts) {
			dports = target.DstPorts
		} else {
			return nil, bosherr.Errorf("Invalid destination port specified %v", target.DstPorts)
		}

		if target.SrcPorts == "" {
			sports = ""
		} else if portPattern.MatchString(target.SrcPorts) {
			sports = target.SrcPorts
		} else {
			return nil, bosherr.Errorf("Invalid destination port specified %v", target.SrcPorts)
		}

		cmd := []string{direction}

		if dsthosts != nil {
			cmd = appendHosts(cmd, "-d", dsthosts...)
		}

		if srchosts != nil {
			cmd = appendHosts(cmd, "-s", srchosts...)
		}

		if protocol != "" {
			cmd = append(cmd, "-p", protocol)
		}

		if dports != "" {
			cmd = append(cmd, "--dport", dports)
		}

		if sports != "" {
			cmd = append(cmd, "--sport", sports)
		}

		cmd = append(cmd, "-j", "DROP")

		rules = append(rules, cmd)
	}

	return rules, nil
}

func (t TargetedBlockerTask) dig(hostname string) ([]string, error) {
	args := []string{"+short", hostname}
	output, _, _, err := t.cmdRunner.RunCommand("dig", args...)
	if err != nil {
		return nil, bosherr.WrapError(err, "resolving host name")
	}

	ips := ipPattern.FindAllString(output, -1)
	if ips == nil {
		return nil, bosherr.Errorf("No IPs found for host %v", hostname)
	}

	return ips, nil
}

func (t TargetedBlockerTask) iptables(action string, rule []string) error {
	args := append([]string{action}, rule...)

	_, _, _, err := t.cmdRunner.RunCommand("iptables", args...)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to iptables")
	}

	return nil
}
