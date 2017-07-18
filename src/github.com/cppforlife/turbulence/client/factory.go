package client

import (
	"fmt"
	"net"
	"net/url"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/http"
	boshhttpclient "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Factory struct {
	logTag string
	logger boshlog.Logger
}

func NewFactory(logger boshlog.Logger) Factory {
	return Factory{
		logTag: "turbulence.Factory",
		logger: logger,
	}
}

func (f Factory) New(config Config) (Turbulence, error) {
	err := config.Validate()
	if err != nil {
		return TurbulenceImpl{}, bosherr.WrapErrorf(err, "Validating Turbulence connection config")
	}

	client, err := f.httpClient(config)
	if err != nil {
		return TurbulenceImpl{}, err
	}

	return TurbulenceImpl{client: client}, nil
}

func (f Factory) httpClient(config Config) (Client, error) {
	certPool, err := config.CACertPool()
	if err != nil {
		return Client{}, err
	}

	if certPool == nil {
		f.logger.Debug(f.logTag, "Using default root CAs")
	} else {
		f.logger.Debug(f.logTag, "Using custom root CAs")
	}

	rawClient := boshhttpclient.CreateDefaultClient(certPool)
	retryClient := boshhttp.NewNetworkSafeRetryClient(rawClient, 5, 500*time.Millisecond, f.logger)
	httpClient := boshhttpclient.NewHTTPClient(retryClient, f.logger)

	endpoint := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(config.Host, fmt.Sprintf("%d", config.Port)),
	}

	return NewClient(endpoint.String(), config.Username, config.Password, httpClient, f.logger), nil
}
