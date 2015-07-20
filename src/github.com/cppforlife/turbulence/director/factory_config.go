package director

import (
	"crypto/x509"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	Host string
	Port int

	// CA certificate is not required
	CACert string

	Username string
	Password string
}

type CPIConfig struct {
	ExePath     string
	JobsDir     string
	PackagesDir string
}

func (c Config) Validate() error {
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

func (c Config) CACertPool() (*x509.CertPool, error) {
	if len(c.CACert) == 0 {
		return nil, nil
	}

	certPool := x509.NewCertPool()

	if ok := certPool.AppendCertsFromPEM([]byte(c.CACert)); !ok {
		return nil, bosherr.Error("Invalid CA certificate")
	}

	return certPool, nil
}

func (c CPIConfig) Exists() bool { return len(c.ExePath) > 0 }

func (c CPIConfig) Validate() error {
	if len(c.ExePath) == 0 {
		return bosherr.Error("Missing 'ExePath'")
	}

	if len(c.JobsDir) == 0 {
		return bosherr.Error("Missing 'JobsDir'")
	}

	if len(c.PackagesDir) == 0 {
		return bosherr.Error("Missing 'PackagesDir'")
	}

	return nil
}
