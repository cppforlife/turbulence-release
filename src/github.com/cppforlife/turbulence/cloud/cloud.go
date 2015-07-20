package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Cloud interface {
	DeleteVM(string) error
}

type cloud struct {
	cpiCmdRunner CPICmdRunner
	context      CmdContext

	logger boshlog.Logger
	logTag string
}

func NewCloud(cpiCmdRunner CPICmdRunner, directorID string, logger boshlog.Logger) Cloud {
	return cloud{
		cpiCmdRunner: cpiCmdRunner,
		context:      CmdContext{DirectorID: directorID},

		logger: logger,
		logTag: "cloud",
	}
}

func (c cloud) DeleteVM(vmCID string) error {
	c.logger.Debug(c.logTag, "Deleting vm '%s'", vmCID)

	method := "delete_vm"

	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method, vmCID)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'delete_vm' method")
	}

	if cmdOutput.Error != nil {
		return NewCPIError(method, *cmdOutput.Error)
	}

	return nil
}
