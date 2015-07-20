package main

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
)

type Factory struct {
	agentID string
	config  APIConfig

	fs        boshsys.FileSystem
	cmdRunner boshsys.CmdRunner

	logTag string
	logger boshlog.Logger
}

func NewFactory(
	agentID string,
	config APIConfig,
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	logger boshlog.Logger,
) Factory {
	return Factory{
		agentID: agentID,
		config:  config,

		fs:        fs,
		cmdRunner: cmdRunner,

		logTag: "agent.Factory",
		logger: logger,
	}
}

func (f Factory) New() (Agent, error) {
	agentConfig, err := f.agentConfig()
	if err != nil {
		return Agent{}, err
	}

	client, err := f.httpClient()
	if err != nil {
		return Agent{}, err
	}

	return newAgent(f.agentID, agentConfig, client, f.cmdRunner, f.logger), nil
}

func (f Factory) agentConfig() (AgentConfig, error) {
	settings, err := NewBOSHSettingsFromPath(f.fs)
	if err != nil {
		return AgentConfig{}, err
	}

	mbusHost, mbusPort, err := settings.HostPort()
	if err != nil {
		return AgentConfig{}, err
	}

	agentConfig := AgentConfig{
		APIHost: f.config.Host,
		APIPort: f.config.Port,

		BOSHMbusHost: mbusHost,
		BOSHMbusPort: mbusPort,
	}

	return agentConfig, nil
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
