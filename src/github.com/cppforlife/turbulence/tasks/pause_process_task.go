package tasks

import(
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type PauseProcessOptions struct {
	Type string
	
	// Times may be suffixed with s,m,h,d,y
	Timeout string

	// Process pattern used with pkill to select what processes are paused.
	ProcessName string
}

func (PauseProcessOptions) _private() {}

type PauseProcessTask struct {
	cmdRunner boshsys.CmdRunner
	opts PauseProcessOptions

	logTag string
	logger boshlog.Logger
}

func NewPauseProcessTask(
	cmdRunner boshsys.CmdRunner,
	opts PauseProcessOptions,
	logger boshlog.Logger,
) PauseProcessTask {
	return PauseProcessTask{cmdRunner, opts, "tasks.PauseProcessTask", logger}
}

func (t PauseProcessTask) Execute(stopCh chan struct{}) error {
	timeoutCh, err := NewMandatoryTimeoutCh(t.opts.Timeout)
	if err != nil {
		return err
	}

	t.logger.Debug(t.logTag, "Pausing processes matching '%s'", t.opts.ProcessName)

	_, _, exitStatus, err := t.cmdRunner.RunCommand("pkill", "-STOP", t.opts.ProcessName)
	if err != nil {
		return bosherr.WrapError(err, "Pausing process")
	} else if exitStatus != 0 {
		return bosherr.Errorf("pkill exited with status %d", exitStatus)
	}

	select {
	case <-timeoutCh:
	case <-stopCh:
	}

	_, _, exitStatus, err = t.cmdRunner.RunCommand("pkill", "-CONT", t.opts.ProcessName)
	if err != nil {
		return bosherr.WrapError(err, "Resuming process")
	} else if exitStatus != 0 {
		return bosherr.Errorf("pkill exited with status %d", exitStatus)
	}

	return nil
}