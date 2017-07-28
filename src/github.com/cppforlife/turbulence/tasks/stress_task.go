package tasks

import (
	"strconv"

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
}

func NewStressTask(cmdRunner boshsys.CmdRunner, opts StressOptions, _ boshlog.Logger) StressTask {
	return StressTask{cmdRunner, opts}
}

func (t StressTask) Execute() error {
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

	if len(t.opts.Timeout) > 0 {
		args = append(args, "--timeout", t.opts.Timeout)
	}

	_, _, _, err := t.cmdRunner.RunCommand("stress", args...)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to stress")
	}

	return nil
}
