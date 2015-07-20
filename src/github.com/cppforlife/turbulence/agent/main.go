package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

var (
	debugOpt      = flag.Bool("debug", false, "Output debug logs")
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	logger, fs, cmdRunner, _ := basicDeps(*debugOpt)
	defer logger.HandlePanic("Main")

	config, err := NewConfigFromPath(*configPathOpt, fs)
	ensureNoErr(logger, "Loading config", err)

	factory := NewFactory(config.AgentID, config.API, fs, cmdRunner, logger)

	agent, err := factory.New()
	ensureNoErr(logger, "Building agent", err)

	err = agent.ContiniouslyExecuteTasks()
	ensureNoErr(logger, "Executing tasks", err)
}

func basicDeps(debug bool) (boshlog.Logger, boshsys.FileSystem, boshsys.CmdRunner, boshuuid.Generator) {
	logLevel := boshlog.LevelInfo

	// Debug generates a lot of log activity
	if debug {
		logLevel = boshlog.LevelDebug
	}

	logger := boshlog.NewWriterLogger(logLevel, os.Stderr, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	runner := boshsys.NewExecCmdRunner(logger)
	uuidGen := boshuuid.NewGenerator()
	return logger, fs, runner, uuidGen
}

func ensureNoErr(logger boshlog.Logger, errPrefix string, err error) {
	if err != nil {
		logger.Error("main", "%s: %s", errPrefix, err)
		os.Exit(1)
	}
}
