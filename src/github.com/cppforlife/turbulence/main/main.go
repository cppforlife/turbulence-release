package main

import (
	"flag"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	mart "github.com/go-martini/martini"
	martauth "github.com/martini-contrib/auth"
	martrend "github.com/martini-contrib/render"

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

	worker := incident.NewWorker(logger)

	scheduler := scheduledinc.NewScheduler(logger)

	go scheduler.Run()

	repos, err := NewRepos(uuidGen, rep, dir, worker, scheduler, logger)
	ensureNoErr(logger, "Failed building repos", err)

	controllerFactory, err := ctrls.NewFactory(repos, logger)
	ensureNoErr(logger, "Failed building controller factory", err)

	err = runControllers(controllerFactory, config, logger)
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

func runControllers(controllerFactory ctrls.Factory, config Config, logger boshlog.Logger) error {
	m := mart.Classic()

	// todo asset is hard coded
	assetsID := "asset-id"

	assetsFuncs := template.FuncMap{
		"cssPath": func(fileName string) (string, error) {
			return "/" + assetsID + "/stylesheets/" + fileName, nil
		},
		"jsPath": func(fileName string) (string, error) {
			return "/" + assetsID + "/javascript/" + fileName, nil
		},
		"imgPath": func(fileName string) (string, error) {
			return "/" + assetsID + "/images/" + fileName, nil
		},
	}

	// Use prefix to cache bust images, stylesheets, and js
	m.Use(mart.Static(
		"./public",
		mart.StaticOptions{
			Prefix: assetsID,
		},
	))

	m.Use(martrend.Renderer(
		martrend.Options{
			Layout:     "layout",
			Directory:  "./templates",
			Extensions: []string{".tmpl", ".html"},
			Funcs:      []template.FuncMap{assetsFuncs},
			IndentJSON: true,
		},
	))

	m.Use(martauth.Basic(config.Username, config.Password))

	m.Get("/", controllerFactory.HomeController.Home)

	isController := controllerFactory.IncidentsController

	m.Get("/incidents", isController.Index)
	m.Get("/incidents/:id", isController.Read)
	m.Get("/api/v1/incidents", isController.APIIndex)
	m.Get("/api/v1/incidents/:id", isController.APIRead)
	m.Post("/api/v1/incidents", isController.APICreate)

	sisController := controllerFactory.ScheduledIncidentsController

	m.Get("/scheduled_incidents", sisController.Index)
	m.Get("/scheduled_incidents/:id", sisController.Read)
	m.Get("/api/v1/scheduled_incidents", sisController.APIIndex)
	m.Get("/api/v1/scheduled_incidents/:id", sisController.APIRead)
	m.Post("/api/v1/scheduled_incidents", sisController.APICreate)
	m.Delete("/api/v1/scheduled_incidents/:id", sisController.APIDelete)

	m.Post("/api/v1/agents/:id/tasks", controllerFactory.AgentRequestsController.APIConsume)
	m.Post("/api/v1/agent_tasks/:id", controllerFactory.AgentRequestsController.APIUpdate)

	return http.ListenAndServeTLS(config.ListenAddr(), config.CertificatePath, config.PrivateKeyPath, m)
}
