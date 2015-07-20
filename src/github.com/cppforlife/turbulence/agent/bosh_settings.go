package main

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type boshSettings struct {
	Mbus string
}

func NewBOSHSettingsFromPath(fs boshsys.FileSystem) (boshSettings, error) {
	var settings boshSettings

	bytes, err := fs.ReadFile("/var/vcap/bosh/settings.json")
	if err != nil {
		return settings, bosherr.WrapErrorf(err, "Reading BOSH settings.json")
	}

	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		return settings, bosherr.WrapError(err, "Unmarshalling BOSH settings")
	}

	return settings, nil
}

func (s boshSettings) HostPort() (string, int, error) {
	mbusURL, err := url.Parse(s.Mbus)
	if err != nil {
		return "", 0, bosherr.WrapError(err, "Parsing BOSH Mbus URL")
	}

	pieces := strings.Split(mbusURL.Host, ":")

	if len(pieces) != 2 {
		return "", 0, bosherr.Errorf("Extracting BOSH Mbus host and port from '%s'", mbusURL.Host)
	}

	port, err := strconv.Atoi(pieces[1])
	if err != nil {
		return "", 0, bosherr.WrapError(err, "Parsing BOSH Mbus port")
	}

	return pieces[0], port, nil
}
