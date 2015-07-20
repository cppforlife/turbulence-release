package director

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cppforlife/turbulence/cloud"
)

type Factory struct {
	config    Config
	cpiConfig CPIConfig
	cmdRunner boshsys.CmdRunner

	logTag string
	logger boshlog.Logger
}

func NewFactory(
	config Config,
	cpiConfig CPIConfig,
	cmdRunner boshsys.CmdRunner,
	logger boshlog.Logger,
) Factory {
	return Factory{
		config:    config,
		cpiConfig: cpiConfig,
		cmdRunner: cmdRunner,

		logTag: "director.Factory",
		logger: logger,
	}
}

func (f Factory) New() (Director, error) {
	client, err := f.httpClient()
	if err != nil {
		return Director{}, err
	}

	cpi, err := f.cpi()
	if err != nil {
		return Director{}, err
	}

	return Director{cpi: cpi, client: client}, nil
}

func (f Factory) httpClient() (Client, error) {
	certPool, err := f.config.CACertPool()
	if err != nil {
		return Client{}, err
	}

	if certPool == nil {
		f.logger.Debug(f.logTag, "Using default root CAs")
	} else {
		f.logger.Debug(f.logTag, "Using custom root CAs")
	}

	httpTransport := &http.Transport{
		TLSClientConfig:     &tls.Config{RootCAs: certPool},
		TLSHandshakeTimeout: 10 * time.Second,

		Dial:  (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 0}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	endpoint := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", f.config.Host, f.config.Port),
		User:   url.UserPassword(f.config.Username, f.config.Password),
	}

	httpClient := boshhttp.NewHTTPClient(http.Client{Transport: httpTransport}, f.logger)

	return NewClient(endpoint.String(), httpClient, f.logger), nil
}

func (f Factory) cpi() (cloud.Cloud, error) {
	if !f.cpiConfig.Exists() {
		return cloud.NewNoopCloud(f.logger), nil
	}

	config := cloud.CPI{
		JobPath:     f.cpiConfig.ExePath,
		JobsDir:     f.cpiConfig.JobsDir,
		PackagesDir: f.cpiConfig.PackagesDir,
	}

	cpiCmdRunner := cloud.NewCPICmdRunner(f.cmdRunner, config, f.logger)

	return cloud.NewCloud(cpiCmdRunner, "fake-director-id", f.logger), nil
}
