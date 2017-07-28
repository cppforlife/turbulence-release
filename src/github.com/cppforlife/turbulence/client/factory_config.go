package client

import (
	"crypto/x509"
	"os"
	"strconv"

	"github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	Host string
	Port int

	Username string
	Password string

	CACert string
}

func NewConfigFromEnv() Config {
	port := 443
	portStr := os.Getenv("TURBULENCE_PORT")

	if len(portStr) > 0 {
		var err error

		port, err = strconv.Atoi(portStr)
		panicIfErr(err, "extract port from env variable")
	}

	config := Config{
		Host: os.Getenv("TURBULENCE_HOST"),
		Port: port,

		Username: os.Getenv("TURBULENCE_USERNAME"),
		Password: os.Getenv("TURBULENCE_PASSWORD"),

		CACert: os.Getenv("TURBULENCE_CA_CERT"),
	}

	return config
}

func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return bosherr.Error("Missing 'Host'")
	}

	if c.Port == 0 {
		return bosherr.Error("Missing 'Port'")
	}

	if len(c.Username) == 0 {
		return bosherr.Error("Missing 'Username'")
	}

	if len(c.Password) == 0 {
		return bosherr.Error("Missing 'Password'")
	}

	if _, err := c.CACertPool(); err != nil {
		return err
	}

	return nil
}

func (c Config) CACertPool() (*x509.CertPool, error) {
	if len(c.CACert) == 0 {
		return nil, nil
	}

	return crypto.CertPoolFromPEM([]byte(c.CACert))
}
