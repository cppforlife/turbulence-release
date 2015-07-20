package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type NoopCloud struct {
	logger boshlog.Logger
	logTag string
}

func NewNoopCloud(logger boshlog.Logger) NoopCloud {
	return NoopCloud{logTag: "cloud", logger: logger}
}

func (c NoopCloud) DeleteVM(vmCID string) error {
	return bosherr.Error("CPI must be configured")
}
