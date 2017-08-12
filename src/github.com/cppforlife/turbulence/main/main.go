package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	ctrls "github.com/cppforlife/turbulence/controllers"
	"github.com/cppforlife/turbulence/director"
	"github.com/cppforlife/turbulence/incident"
	"github.com/cppforlife/turbulence/incident/reporter"
	"github.com/cppforlife/turbulence/scheduledinc"
)

const mainLogTag = "main"

var (
	debugOpt      = flag.Bool("debug", false, "Output debug logs")
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	logger, fs, uuidGen := basicDeps(*debugOpt)
	defer logger.HandlePanic("Main")

	config, err := NewConfigFromPath(*configPathOpt, fs)
	ensureNoErr(logger, "Loading config", err)

	var rep reporter.Reporter

	{
		if config.Datadog.Required() {
			rep = reporter.NewDatadog(config.Datadog, logger)
		} else {
			rep = reporter.NewLogger(logger)
		}
	}

	var dir director.Director

	{
		directorFactory := director.NewFactory(config.Director, logger)

		dir, err = directorFactory.New()
		ensureNoErr(logger, "Failed building director", err)
	}

	rep = reporter.NewMulti([]reporter.Reporter{reporter.NewDirectorEvents(dir, logger), rep})

	worker := incident.NewWorker(logger)

	scheduler := scheduledinc.NewScheduler(logger)

	go scheduler.Run()

	repos, err := NewRepos(uuidGen, rep, dir, worker, scheduler, logger)
	ensureNoErr(logger, "Failed building repos", err)

	controllerFactory, err := ctrls.NewFactory(repos, logger)
	ensureNoErr(logger, "Failed building controller factory", err)

	err = Server{config, logger}.RunControllers(controllerFactory)
	ensureNoErr(logger, "Running controllers", err)
}

func basicDeps(debug bool) (boshlog.Logger, boshsys.FileSystem, boshuuid.Generator) {
	logLevel := boshlog.LevelInfo

	// Debug generates a lot of log activity
	if debug {
		logLevel = boshlog.LevelDebug
	}

	logger := boshlog.NewWriterLogger(logLevel, os.Stderr, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	uuidGen := boshuuid.NewGenerator()
	return logger, fs, uuidGen
}

func ensureNoErr(logger boshlog.Logger, errPrefix string, err error) {
	if err != nil {
		logger.Error(mainLogTag, "%s: %s", errPrefix, err)
		os.Exit(1)
	}
}
