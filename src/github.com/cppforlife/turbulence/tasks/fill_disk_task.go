package tasks

import (
	// bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FillDiskOptions struct {
	Type string
	Timeout string

	// todo to percentage

	// By default root disk will be filled
	Persistent bool
	Ephemeral  bool
	Temporary  bool
}

func (FillDiskOptions) _private() {}

type FillDiskTask struct {
	cmdRunner boshsys.CmdRunner
	opts      FillDiskOptions

	logTag string
	logger boshlog.Logger
}

func NewFillDiskTask(cmdRunner boshsys.CmdRunner, opts FillDiskOptions, logger boshlog.Logger) FillDiskTask {
	return FillDiskTask{cmdRunner, opts, "tasks.FillDiskTask", logger}
}

func (t FillDiskTask) Execute(stopCh chan struct{}) error {
	timeoutCh, err := NewOptionalTimeoutCh(t.opts.Timeout)
	if err != nil {
		return err
	}

	if t.opts.Persistent {
		err = t.fill("/var/vcap/store/.filler")
	} else if t.opts.Ephemeral {
		err = t.fill("/var/vcap/data/.filler")
	} else if t.opts.Temporary {
		err = t.fill("/tmp/.filler")
	} else {
		err = t.fill("/.filler")
	}

	if err != nil {
		return err
	}

	select {
	case <-stopCh:
	case <-timeoutCh:
	}

	if t.opts.Persistent {
		err = t.remove("/var/vcap/store/.filler")
	} else if t.opts.Ephemeral {
		err = t.remove("/var/vcap/data/.filler")
	} else if t.opts.Temporary {
		err = t.remove("/tmp/.filler")
	} else {
		err = t.remove("/.filler")
	}

	return err
}

func (t FillDiskTask) fill(path string) error {
	_, _, _, err := t.cmdRunner.RunCommand("dd", "if=/dev/zero", "of="+path, "bs=1M")
	if err != nil {
		t.logger.Debug(t.logTag, "Encountered error filling disk: ", err)
	}
	// don't stop because of an error because it is probably from it running out of disk space which is to be expected
	
	return nil
}

func (t FillDiskTask) remove(path string) error {
	_, _, _, err := t.cmdRunner.RunCommand("rm", path)
	return err
}
