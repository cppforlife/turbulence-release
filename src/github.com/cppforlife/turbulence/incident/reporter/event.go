package reporter

import (
	"sync"
	"time"
)

const (
	EventTypeFind   = "Find"
	EventTypeSelect = "Select"
)

type Event struct {
	reporter   Reporter
	incidentID string

	resultsWg *sync.WaitGroup

	ID   string // may be empty
	Type string

	Instance EventInstance // may be empty

	ExecutionStartedAt   time.Time
	ExecutionCompletedAt time.Time

	Error error
}

type EventInstance struct {
	ID         string
	Group      string
	Deployment string
	AZ         string
}

func (e *Event) IsAction() bool {
	return e.Type != EventTypeFind && e.Type != EventTypeSelect
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
