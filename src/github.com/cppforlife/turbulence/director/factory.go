package director

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Factory struct {
	config Config
	logger boshlog.Logger
}

func NewFactory(config Config, logger boshlog.Logger) Factory {
	return Factory{config: config, logger: logger}
}

func (c Factory) New() (Director, error) {
	director, err := c.director()
	if err != nil {
		return nil, err
	}

	return DirectorImpl{director}, nil
}

func (c Factory) director() (boshdir.Director, error) {
	info, err := c.info()
	if err != nil {
		return nil, err
	}

	dirConfig := c.config.UserConfig()

	if info.Auth.Type == "uaa" {
		uaa, err := c.uaa(info)
		if err != nil {
			return nil, err
		}

		dirConfig.Client = ""
		dirConfig.ClientSecret = ""

		dirConfig.TokenFunc = boshuaa.NewClientTokenSession(uaa).TokenFunc
	}

	taskReporter := boshdir.NewNoopTaskReporter()
	fileReporter := boshdir.NewNoopFileReporter()

	return boshdir.NewFactory(c.logger).New(dirConfig, taskReporter, fileReporter)
}

func (c Factory) uaa(info boshdir.Info) (boshuaa.UAA, error) {
	uaaURL := info.Auth.Options["url"]

	uaaURLStr, ok := uaaURL.(string)
	if !ok {
		return nil, bosherr.Errorf("Expected URL '%s' to be a string", uaaURL)
	}

	uaaConfig, err := boshuaa.NewConfigFromURL(uaaURLStr)
	if err != nil {
		return nil, err
	}

	uaaConfig.CACert = c.config.CACert
	uaaConfig.Client = c.config.Client
	uaaConfig.ClientSecret = c.config.ClientSecret

	if len(uaaConfig.Client) == 0 {
		uaaConfig.Client = "bosh_cli"
	}

	return boshuaa.NewFactory(c.logger).New(uaaConfig)
}

func (c Factory) info() (boshdir.Info, error) {
	dirConfig := c.config.AnonymousUserConfig()

	director, err := boshdir.NewFactory(c.logger).New(dirConfig, nil, nil)
	if err != nil {
		return boshdir.Info{}, err
	}

	return director.Info()
}
