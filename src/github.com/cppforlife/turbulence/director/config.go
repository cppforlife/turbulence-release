package director

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	Host string
	Port int

	CACert string

	Client       string
	ClientSecret string
}

func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return bosherr.Error("Missing 'Host'")
	}

	if c.Port == 0 {
		return bosherr.Error("Missing 'Port'")
	}

	if len(c.Client) == 0 {
		return bosherr.Error("Missing 'Client'")
	}

	if len(c.ClientSecret) == 0 {
		return bosherr.Error("Missing 'ClientSecret'")
	}

	return nil
}

func (c Config) AnonymousUserConfig() boshdir.Config {
	return boshdir.Config{
		Host: c.Host,
		Port: c.Port,

		CACert: c.CACert,
	}
}

func (c Config) UserConfig() boshdir.Config {
	return boshdir.Config{
		Host: c.Host,
		Port: c.Port,

		CACert: c.CACert,

		Client:       c.Client,
		ClientSecret: c.ClientSecret,
	}
}
