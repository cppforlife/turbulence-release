package reporter

import (
	"time"
)

type Incident interface {
	ID() string

	TaskTypes() []string
	ShortDescription() (string, error)

	Events() *Events

	ExecutionStartedAt() time.Time
	ExecutionCompletedAt() time.Time
}

type Reporter interface {
	ReportIncidentExecutionStart(Incident)
	ReportIncidentExecutionCompletion(Incident)

	ReportEventExecutionStart(string, Event)
	ReportEventExecutionCompletion(string, Event)
}

var _ Reporter = Multi{}
var _ Reporter = Logger{}
var _ Reporter = DirectorEvents{}
