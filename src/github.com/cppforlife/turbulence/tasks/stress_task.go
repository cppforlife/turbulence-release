package tasks

import (
	"strconv"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type StressOptions struct {
	Type    string
	Timeout string // Times may be suffixed with s,m,h,d,y

	NumCPUWorkers int
	NumIOWorkers  int

	NumMemoryWorkers  int
	MemoryWorkerBytes string // Sizes may be suffixed with B,K,M,G

	NumHDDWorkers  int
	HDDWorkerBytes string // Sizes may be suffixed with B,K,M,G
}

func (StressOptions) _private() {}

type StressTask struct {
	cmdRunner boshsys.CmdRunner
	opts      StressOptions

	logTag string
	logger boshlog.Logger
}

func NewStressTask(cmdRunner boshsys.CmdRunner, opts StressOptions, logger boshlog.Logger) StressTask {
	return StressTask{cmdRunner, opts, "task.StressTask", logger}
}

func (t StressTask) Execute(stopCh chan struct{}) error {
	// e.g. stress --cpu 2 --io 1 --vm 1 --vm-bytes 128M --timeout 10s --verbose

	args := []string{"--verbose"}

	if t.opts.NumCPUWorkers+t.opts.NumIOWorkers+t.opts.NumMemoryWorkers+t.opts.NumHDDWorkers == 0 {
		return bosherr.Error("Must specify at least 1 type of worker")
	}

	if t.opts.NumCPUWorkers > 0 {
		args = append(args, "--cpu", strconv.Itoa(t.opts.NumCPUWorkers))
	}

	if t.opts.NumIOWorkers > 0 {
		args = append(args, "--io", strconv.Itoa(t.opts.NumIOWorkers))
	}

	if t.opts.NumMemoryWorkers > 0 {
		if len(t.opts.MemoryWorkerBytes) == 0 {
			return bosherr.Error("Must specify 'MemoryWorkerBytes'")
		}

		args = append(
			args,
			"--vm", strconv.Itoa(t.opts.NumMemoryWorkers),
			"--vm-bytes", t.opts.MemoryWorkerBytes,
		)
	}

	if t.opts.NumHDDWorkers > 0 {
		if len(t.opts.HDDWorkerBytes) == 0 {
			return bosherr.Error("Must specify 'HDDWorkerBytes'")
		}

		args = append(
			args,
			"--hdd", strconv.Itoa(t.opts.NumHDDWorkers),
			"--hdd-bytes", t.opts.HDDWorkerBytes,
		)
	}

	// todo remove timeout option?
	if len(t.opts.Timeout) > 0 {
		args = append(args, "--timeout", t.opts.Timeout)
	}

	return t.runStress(args, stopCh)
}

func (t StressTask) runStress(args []string, stopCh chan struct{}) error {
	command := boshsys.Command{
		Name: "stress",
		Args: args,
	}

	process, err := t.cmdRunner.RunComplexCommandAsync(command)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to stress")
	}

	var result boshsys.Result

	isStopped := false

	// Can only wait once on a process but cancelling can happen multiple times
	for procExitedCh := process.Wait(); procExitedCh != nil; {
		select {
		case result = <-procExitedCh:
			procExitedCh = nil
		case <-stopCh:
			// Ignore possible TerminateNicely error since we cannot return it
			err := process.TerminateNicely(10 * time.Second)
			if err != nil {
				t.logger.Error(t.logTag, "Failed to terminate %s", err.Error())
			}
			isStopped = true
		}
	}

	if isStopped {
		return nil // todo successfully stopped?
	}

	if result.Error != nil {
		return bosherr.WrapError(result.Error, "Running stress")
	}

	return nil
}
