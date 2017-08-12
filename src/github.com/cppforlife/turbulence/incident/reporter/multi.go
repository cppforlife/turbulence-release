package reporter

type Multi struct {
	reps []Reporter
}

func NewMulti(reps []Reporter) Multi {
	return Multi{reps}
}

func (r Multi) ReportIncidentExecutionStart(i Incident) {
	for _, rep := range r.reps {
		rep.ReportIncidentExecutionStart(i)
	}
}

func (r Multi) ReportIncidentExecutionCompletion(i Incident) {
	for _, rep := range r.reps {
		rep.ReportIncidentExecutionCompletion(i)
	}
}

func (r Multi) ReportEventExecutionStart(incidentID string, e Event) {
	for _, rep := range r.reps {
		rep.ReportEventExecutionStart(incidentID, e)
	}
}

func (r Multi) ReportEventExecutionCompletion(incidentID string, e Event) {
	for _, rep := range r.reps {
		rep.ReportEventExecutionCompletion(incidentID, e)
	}
}
