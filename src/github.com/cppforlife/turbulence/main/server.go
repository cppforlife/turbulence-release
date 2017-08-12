package main

import (
	"html/template"
	"net/http"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	mart "github.com/go-martini/martini"
	martauth "github.com/martini-contrib/auth"
	martrend "github.com/martini-contrib/render"

	ctrls "github.com/cppforlife/turbulence/controllers"
)

type Server struct {
	config Config
	logger boshlog.Logger
}

func (s Server) RunControllers(controllerFactory ctrls.Factory) error {
	operatorM := s.authedMartini()
	s.addOperatorAPI(operatorM, controllerFactory)

	agentM := s.authedMartini()
	s.addAgentAPI(agentM, controllerFactory)

	errs := make(chan error, 1)
	go s.listen(s.config.ListenAddr(), "operator", operatorM, errs)
	go s.listen(s.config.AgentListenAddr(), "agent", agentM, errs)

	return <-errs
}

func (s Server) listen(addr string, purpose string, m *mart.ClassicMartini, errs chan<- error) {
	s.logger.Debug("main.Server", "Starting %s API '%s'", purpose, addr)
	errs <- http.ListenAndServeTLS(addr, s.config.CertificatePath, s.config.PrivateKeyPath, m)
}

func (s Server) authedMartini() *mart.ClassicMartini {
	m := mart.Classic()
	m.Use(martauth.Basic(s.config.Username, s.config.Password))
	return m
}

func (s Server) addOperatorAPI(m *mart.ClassicMartini, controllerFactory ctrls.Factory) {
	s.addAssets(m)

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

	// Driver may change desired state of the task so that task ends
	m.Post("/api/v1/agent_tasks/:id/state", controllerFactory.TasksController.APIUpdateState)
}

func (Server) addAgentAPI(m *mart.ClassicMartini, controllerFactory ctrls.Factory) {
	m.Use(martrend.Renderer(martrend.Options{
		Layout:     "",
		Directory:  "./templates", // just in case
		Extensions: []string{".ext-does-not-exist"},
		IndentJSON: true,
	}))

	// Agent watches for tasks based on agent ID
	m.Post("/api/v1/agents/:id/tasks", controllerFactory.TasksController.APIConsume)
	// Agent watches desired state of the task so that it can end it
	m.Get("/api/v1/agent_tasks/:id/state", controllerFactory.TasksController.APIReadState)
	// Once agent executes picked up task, its result is reported
	m.Post("/api/v1/agent_tasks/:id", controllerFactory.TasksController.APIUpdate)
}

func (Server) addAssets(m *mart.ClassicMartini) {
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
	m.Use(mart.Static("./public", mart.StaticOptions{Prefix: assetsID}))

	m.Use(martrend.Renderer(martrend.Options{
		Layout:     "layout",
		Directory:  "./templates",
		Extensions: []string{".tmpl", ".html"},
		Funcs:      []template.FuncMap{assetsFuncs},
		IndentJSON: true,
	}))
}
