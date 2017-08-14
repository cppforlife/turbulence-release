package main

import (
	"encoding/json"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cppforlife/turbulence/director"
	"github.com/cppforlife/turbulence/incident/reporter"
)

type Config struct {
	ListenAddress   string
	ListenPort      int
	AgentListenPort int

	Username string
	Password string

	CertificatePath string
	PrivateKeyPath  string

	Director director.Config

	Datadog reporter.DatadogConfig
}

func NewConfigFromPath(path string, fs boshsys.FileSystem) (Config, error) {
	var config Config

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return config, bosherr.WrapErrorf(err, "Reading config %s", path)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config")
	}

	err = config.Validate()
	if err != nil {
		return config, bosherr.WrapError(err, "Validating config")
	}

	return config, nil
}

func (c Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.ListenAddress, c.ListenPort)
}

func (c Config) AgentListenAddr() string {
	return fmt.Sprintf("%s:%d", c.ListenAddress, c.AgentListenPort)
}

func (c Config) Validate() error {
	if len(c.ListenAddress) == 0 {
		return bosherr.Error("Missing 'ListenAddress'")
	}

	if c.ListenPort == 0 {
		return bosherr.Error("Missing 'ListenPort'")
	}

	if c.AgentListenPort == 0 {
		return bosherr.Error("Missing 'AgentListenPort'")
	}

	if len(c.Username) == 0 {
		return bosherr.Error("Missing 'Username'")
	}

	if len(c.Password) == 0 {
		return bosherr.Error("Missing 'Password'")
	}

	if len(c.CertificatePath) == 0 {
		return bosherr.Error("Missing 'CertificatePath'")
	}

	if len(c.PrivateKeyPath) == 0 {
		return bosherr.Error("Missing 'PrivateKeyPath'")
	}

	err := c.Director.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating 'Director' config")
	}

	err = c.Datadog.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating 'Datadog' config")
	}

	return nil
}
