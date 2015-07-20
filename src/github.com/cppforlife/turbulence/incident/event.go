package incident

import (
	"sync"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

const (
	EventTypeFindDeployment  = "FindDeployment"
	EventTypeFindJobs        = "FindJobs"
	EventTypeFindInstances   = "FindInstances"
	EventTypeSelectInstances = "SelectInstances"
)

type Events struct {
	uuidGen    boshuuid.Generator
	reporter   Reporter
	incidentID string

	resultsWg sync.WaitGroup
	resultsCh chan EventResult

	events []*Event

	logger boshlog.Logger
}

type Event struct {
	reporter   Reporter
	incidentID string

	resultsWg *sync.WaitGroup

	ID   string // may be empty
	Type string

	DeploymentName string
	JobName        string
	JobNameMatch   string
	JobIndex       *int

	ExecutionStartedAt   time.Time
	ExecutionCompletedAt time.Time

	Error error
}

type EventResult struct {
	Event *Event
	Error error
}

func NewEvents(uuidGen boshuuid.Generator, reporter Reporter, incidentID string, logger boshlog.Logger) *Events {
	return &Events{
		uuidGen:    uuidGen,
		reporter:   reporter,
		incidentID: incidentID,
		resultsCh:  make(chan EventResult),
		logger:     logger,
	}
}

func (e *Events) RegisterResult(r EventResult) {
	e.resultsCh <- r
}

func (e *Events) Results() chan EventResult {
	go func() {
		e.resultsWg.Wait()
		close(e.resultsCh)
	}()

	return e.resultsCh
}

func (e *Events) Add(event Event) *Event {
	e.events = append(e.events, &event)

	e.resultsWg.Add(1)

	event.resultsWg = &e.resultsWg
	event.reporter = e.reporter

	id, err := e.uuidGen.Generate()
	if err != nil {
		// Allow event ID to be an empty string
		e.logger.Error("Events", "Failed to generate event ID")
	}

	event.ID = id
	event.incidentID = e.incidentID
	event.ExecutionStartedAt = time.Now().UTC()

	e.reporter.ReportEventExecutionStart(e.incidentID, event)

	return &event
}

func (e *Events) Events() []*Event {
	if e == nil {
		return nil
	}

	return e.events
}

func (e *Events) FirstError() error {
	for _, ev := range e.events {
		if ev.Error != nil {
			return ev.Error
		}
	}

	return nil
}

func (e *Event) ErrorStr() string {
	if e.Error != nil {
		return e.Error.Error()
	}

	return ""
}

func (e *Event) MarkError(err error) bool {
	e.Error = err
	e.ExecutionCompletedAt = time.Now().UTC()

	// Call it done after updating event
	e.resultsWg.Done()

	e.reporter.ReportEventExecutionCompletion(e.incidentID, *e)

	return err != nil
}

func (e *Event) IsAction() bool {
	return e.Type != EventTypeFindDeployment &&
		e.Type != EventTypeFindJobs &&
		e.Type != EventTypeFindInstances &&
		e.Type != EventTypeSelectInstances
}
