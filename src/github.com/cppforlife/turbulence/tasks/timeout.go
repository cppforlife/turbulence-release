package tasks

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func NewOptionalTimeoutCh(timeoutStr string) (<-chan time.Time, error) {
	return newTimeoutCh(timeoutStr, true)
}

func NewMandatoryTimeoutCh(timeoutStr string) (<- chan time.Time, error) {
	return newTimeoutCh(timeoutStr, false)
}

func newTimeoutCh(timeoutStr string, optional bool) (<-chan time.Time, error) {
	if optional && len(timeoutStr) == 0 {
		return nil, bosherr.Error("Timeout must be specified.")
	}

	if len(timeoutStr) == 0 {
		return make(chan time.Time), nil // never fires
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing timeout")
	}

	return time.After(timeout), nil
}