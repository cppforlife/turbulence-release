package monit

import (
	boshhttp "github.com/cloudfoundry/bosh-utils/http"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type httpClient struct {
	host, username, password string
	client                   boshhttp.Client
	logger                   boshlog.Logger
}

func NewHTTPClient(
	host, username, password string,
	client boshhttp.Client,
	logger boshlog.Logger,
) Client {
	return httpClient{
		host:     host,
		username: username,
		password: password,
		client:   client,
		logger:   logger,
	}
}
