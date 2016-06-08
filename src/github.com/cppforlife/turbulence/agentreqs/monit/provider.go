package monit

import (
	"net/http"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/http"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

var (
	monitCredsPath = "/var/vcap/monit/monit.user"
	monitHost      = "127.0.0.1:2822"
)

type ClientProvider struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewClientProvider(fs boshsys.FileSystem, logger boshlog.Logger) ClientProvider {
	return ClientProvider{fs: fs, logger: logger}
}

func (p ClientProvider) Get() (Client, error) {
	credsStr, err := p.fs.ReadFileString(monitCredsPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting monit credentials")
	}

	creds := strings.SplitN(credsStr, ":", 2)

	httpClient := boshhttp.NewRetryClient(http.DefaultClient, 2, 1*time.Second, p.logger)

	return NewHTTPClient(monitHost, creds[0], creds[1], httpClient, p.logger), nil
}
