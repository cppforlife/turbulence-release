package agentreqs

import (
	"net"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func NonLocalIfaceNames() ([]string, error) {
	var ifaceNames []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return ifaceNames, bosherr.WrapError(err, "Listing network interfaces")
	}

	for _, iface := range ifaces {
		if !strings.HasPrefix(iface.Name, "lo") {
			ifaceNames = append(ifaceNames, iface.Name)
		}
	}

	return ifaceNames, nil
}
