package tasks

import (
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

// See http://www.linuxfoundation.org/collaborate/workgroups/networking/netem
// See http://mark.koli.ch/slowdown-throttle-bandwidth-linux-network-interface
type ControlNetOptions struct {
	Type    string
	Timeout string // Times may be suffixed with ms,s,m,h

	// slow: tc qdisc add dev eth0 root netem delay 50ms 10ms distribution normal
	Delay          string
	DelayVariation string

	// flaky: tc qdisc add dev eth0 root netem loss 20% 75%
	Loss            string
	LossCorrelation string

	// tc qdisc add dev eth0 root netem duplicate 1%
	Duplication string

	// tc qdisc add dev eth0 root netem corrupt 0.1%
	Corruption string

	// tc qdisc add dec eth0 root netem reorder 25% 50%
	Reorder            string
	ReorderCorrelation string

	// tc qdisc add dev eth0 root handle 1:0 [netem...]
	// tc qdisc add dev eth0 parent 1:1 handle 10: tfb rate 256kbit buffer 1600 limit 3000
	Bandwidth string

	// reset: tc qdisc del dev eth0 root
	
	// since tc is for egress only, specify which targets to affect
	Targets []DestinationTarget
}

type DestinationTarget struct {
	// Optional destination host to block, can specify an address such as "10.34.4.60", an address block such as "192.168.0.0/24",
	// or a domain name such as "google.com" which will be resolved to an Ip.
	DstHost string


	// Optional "dport" or destination port(s) to block. No range of ports is supported, as this is too dificult to implement via masking: https://serverfault.com/questions/231880/how-to-match-port-range-using-u32-filter
	DstPort string
}

func (ControlNetOptions) _private() {}

type ControlNetTask struct {
	cmdRunner boshsys.CmdRunner
	opts      ControlNetOptions
}

func NewControlNetTask(cmdRunner boshsys.CmdRunner, opts ControlNetOptions, _ boshlog.Logger) ControlNetTask {
	return ControlNetTask{cmdRunner, opts}
}

func defaultStr(v, d string) string {
	if len(v) == 0 {
		return d
	} else {
		return v
	}
}

func (t ControlNetTask) Execute(stopCh chan struct{}) error {
	timeoutCh, err := NewOptionalTimeoutCh(t.opts.Timeout)
	if err != nil {
		return err
	}

	delay, loss, duplication, corruption, reorder, bandwidth := len(t.opts.Delay) > 0, len(t.opts.Loss) > 0, len(t.opts.Duplication) > 0, len(t.opts.Corruption) > 0, len(t.opts.Reorder) > 0, len(t.opts.Bandwidth) > 0

	if !(delay || loss || duplication || corruption || reorder || bandwidth) {
		return bosherr.Error("Must specify an effect")
	}

	if bandwidth && (delay || loss || duplication || corruption || reorder) {
		return bosherr.Error("Cannot limit the bandwidth at the same time as other effects")
	}

	ifaceNames, err := NonLocalIfaceNames()
	if err != nil {
		return err
	}

	if bandwidth {
		for _, ifaceName := range ifaceNames {
			err := t.configureBandwidth(ifaceName)
			if err != nil {
				t.resetIfaces(ifaceNames)
				return err
			}
		}
	} else {
		opts := make([]string, 0, 16)

		if delay {
			variation := defaultStr(t.opts.DelayVariation, "10ms")
			opts = append(opts, "delay", t.opts.Delay, variation, "distribution", "normal")
		}

		if loss {
			correlation := defaultStr(t.opts.LossCorrelation, "75%")
			opts = append(opts, "loss", t.opts.Loss, correlation)
		}

		if duplication {
			opts = append(opts, "duplicate", t.opts.Duplication)
		}

		if corruption {
			opts = append(opts, "corrupt", t.opts.Corruption)
		}

		if reorder {
			correlation := defaultStr(t.opts.ReorderCorrelation, "50%")
			opts = append(opts, "reorder", t.opts.Reorder, correlation)
		}

		for _, ifaceName := range ifaceNames {
			err := t.configureInterface(ifaceName, opts)
			if err != nil {
				t.resetIfaces(ifaceNames)
				return err
			}
		}
	}

	select {
	case <-timeoutCh:
	case <-stopCh:
	}


	return t.resetIfaces(ifaceNames)
}

func (t ControlNetTask) configureInterface(ifaceName string, opts []string) error {
	_, _, _, err := t.cmdRunner.RunCommand("tc", "qdisc", "add", "dev", ifaceName, "root", "handle", "1:", "prio")
	if err != nil {
		return err
	}

	args := []string{"qdisc", "add", "dev", ifaceName, "parent", "1:1", "handle", "30:", "netem"}
	args = append(args, opts...)
	_, _, _, err = t.cmdRunner.RunCommand("tc", args...)
	if err != nil {
		return err
	}

	return t.configureDestination(ifaceName)
}

func (t ControlNetTask) configureBandwidth(ifaceName string) error {
	_, _, _, err := t.cmdRunner.RunCommand("tc", "qdisc", "add", "dev", ifaceName, "root", "handle", "1:", "htb")
	if err != nil {
		return err
	}

	_, _, _, err = t.cmdRunner.RunCommand("tc", "class", "add", "dev", ifaceName, "parent", "1:", "classid", "1:1", "htb", "rate", t.opts.Bandwidth)
	if err != nil {
		return err
	}

	return t.configureDestination(ifaceName)
}

func (t ControlNetTask) configureDestination(ifaceName string) error {
	rules := [][]string{}

	if len(t.opts.Targets) == 0 {
		// we need to add this to forward the traffic to the default class
		rules = [][]string{[]string{"match", "ip", "dst", "0.0.0.0/0"}}
	} else {
		for _, target := range t.opts.Targets {
			if target.DstHost == "" && target.DstPort == "" {
				return bosherr.Error("Must specify at least one of DstHost or DstPort.")
			}

			var dsthosts []string
			var dport string

			if target.DstPort == "" {
				dport = ""
			} else if destinationPortPattern.MatchString(target.DstPort) {
				dport = target.DstPort
			} else {
				return bosherr.Errorf("Invalid destination port specified %v", target.DstPort)
			}

			dsthosts, err := t.getHost(target.DstHost)
			if err != nil {
				return err
			}
			
			if len(dsthosts) == 0 {
				// only port was specified
				rules = append(rules, []string{"match", "ip", "dport", dport, "0xffff"})
			} else {
				for _, dsthost := range dsthosts {
					args := []string{"match", "ip", "dst", dsthost}
					
					if (dport != "") {
						// check if we have to add the port to the same rule
						args = append(args, []string{"match", "ip", "dport", dport, "0xffff"}...)
					}
					
					rules = append(rules, args)
				}
			}
		}
	}
	
	for _, rule := range rules {
		args := []string{"filter", "add", "dev", ifaceName, "protocol", "ip", "parent", "1:", "prio", "1", "u32"}
		args = append(args, rule...)
		args = append(args, []string{"flowid", "1:1"}...)

		_, _, _, err := t.cmdRunner.RunCommand("tc", args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t ControlNetTask) resetIfaces(ifaceNames []string) error {
	errors := []error{}
	for _, ifaceName := range ifaceNames {
		_, _, _, err := t.cmdRunner.RunCommand("tc", "qdisc", "del", "dev", ifaceName, "root")
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		msgs := []string{}
		for _, error := range errors {
			msgs = append(msgs, error.Error())
		}
		
		return bosherr.Errorf("Errors detected during reset: %s", strings.Join(msgs," "))
	}

	return nil
}

var destinationIpPattern = regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(/\d{0,2})?`)
var destinationPortPattern = regexp.MustCompile(`\d+(:\d+)?$`)

func (t ControlNetTask) dig(hostname string) ([]string, error) {
	args := []string{"+short", hostname}
	output, _, _, err := t.cmdRunner.RunCommand("dig", args...)
	if err != nil {
		return nil, bosherr.WrapError(err, "resolving host name")
	}

	ips := destinationIpPattern.FindAllString(output, -1)
	if ips == nil {
		return nil, bosherr.Errorf("No IPs found for host %v", hostname)
	}

	return ips, nil
}

func (t ControlNetTask) getHost(host string) ([]string, error) {
	if host == "" {
		return nil, nil
	} else if destinationIpPattern.MatchString(host) {
		return destinationIpPattern.FindAllString(host, -1), nil
	} else {
		return t.dig(host)
	}
}
