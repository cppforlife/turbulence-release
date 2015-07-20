package main

import (
	"crypto/x509"
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Config struct {
	AgentID string

	API APIConfig
}

type APIConfig struct {
	Host string
	Port int

	// CA certificate is not required
	CACert string

	Username string
	Password string
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

func (c Config) Validate() error {
	err := c.API.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating 'API' config")
	}

	return nil
}

func (c APIConfig) CACertPool() (*x509.CertPool, error) {
	if len(c.CACert) == 0 {
		return nil, nil
	}

	certPool := x509.NewCertPool()

	if ok := certPool.AppendCertsFromPEM([]byte(c.CACert)); !ok {
		return nil, bosherr.Error("Invalid CA certificate")
	}

	return certPool, nil
}

func (c APIConfig) Validate() error {
	if len(c.Host) == 0 {
		return bosherr.Error("Missing 'Host'")
	}

	if c.Port == 0 {
		return bosherr.Error("Missing 'Port'")
	}

	if _, err := c.CACertPool(); err != nil {
		return err
	}

	if len(c.Username) == 0 {
		return bosherr.Error("Missing 'Username'")
	}

	if len(c.Password) == 0 {
		return bosherr.Error("Missing 'Password'")
	}

	return nil
}
