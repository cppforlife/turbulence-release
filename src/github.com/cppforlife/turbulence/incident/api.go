package incident

import (
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/turbulence/agentreqs"
)

const (
	IncidentTypeKill         = "kill"
	IncidentTypeStress       = "stress"       // cpu, ram, io
	IncidentTypeControlNet   = "control-net"  // drop X% of traffic, restrict X% bw, increase latency
	IncidentTypeFillStorage  = "fill-storage" // fill up ephemeral/root/store disk
	IncidentTypeKillProcess  = "kill-process"
	IncidentTypePauseProcess = "pause-process"
)

type IncidentReq struct {
	Tasks       agentreqs.TaskOptionsSlice
	Deployments []Deployment
}

type Deployment struct {
	Name string
	Jobs []Job
}

type Job struct {
	Name string

	Indices []int
	Limit   string
}

type IncidentResp struct {
	incident Incident

	ID string

	Tasks       agentreqs.TaskOptionsSlice
	Deployments []Deployment

	ExecutionStartedAt   string
	ExecutionCompletedAt string

	Events []EventResp

	description string
}

type IncidentsResp []IncidentResp

type EventResp struct {
	event *Event

	ID   string
	Type string

	DeploymentName string
	JobName        string
	JobNameMatch   string
	JobIndex       *int

	ExecutionStartedAt   string
	ExecutionCompletedAt string

	Error string
}

func NewIncidentsResp(incidents []Incident) IncidentsResp {
	resp := []IncidentResp{}

	for _, incid := range incidents {
		resp = append(resp, NewIncidentResp(incid))
	}

	return resp
}

func NewIncidentResp(incident Incident) IncidentResp {
	var eventResps []EventResp

	for _, event := range incident.Events.Events() {
		eventResps = append(eventResps, NewEventResp(event))
	}

	var completedAt string

	if (incident.ExecutionCompletedAt != time.Time{}) {
		completedAt = incident.ExecutionCompletedAt.Format(time.RFC3339)
	}

	return IncidentResp{
		incident: incident,

		ID: incident.ID,

		Tasks:       incident.Tasks,
		Deployments: incident.Deployments,

		ExecutionStartedAt:   incident.ExecutionStartedAt.Format(time.RFC3339),
		ExecutionCompletedAt: completedAt,

		Events: eventResps,
	}
}

func NewEventResp(event *Event) EventResp {
	var completedAt string

	if (event.ExecutionCompletedAt != time.Time{}) {
		completedAt = event.ExecutionCompletedAt.Format(time.RFC3339)
	}

	return EventResp{
		event: event,

		ID:   event.ID,
		Type: event.Type,

		DeploymentName: event.DeploymentName,
		JobName:        event.JobName,
		JobNameMatch:   event.JobNameMatch,
		JobIndex:       event.JobIndex,

		ExecutionStartedAt:   event.ExecutionStartedAt.Format(time.RFC3339),
		ExecutionCompletedAt: completedAt,

		Error: event.ErrorStr(),
	}
}

func (r IncidentResp) URL() string { return fmt.Sprintf("/incidents/%s", r.ID) }

func (r IncidentResp) TaskTypes() string { return strings.Join(r.incident.TaskTypes(), ", ") }

func (r IncidentResp) Description() (string, error) { return r.incident.Description() }

func (r IncidentResp) HasEventErrors() bool {
	for _, eventResp := range r.Events {
		if len(eventResp.Error) > 0 {
			return true
		}
	}

	return false
}

func (r EventResp) IsAction() bool { return r.event.IsAction() }

func (r EventResp) DescriptionHTML() template.HTML {
	var descPieces []string

	if len(r.DeploymentName) > 0 && len(r.JobName) > 0 && r.JobIndex != nil {
		return template.HTML(fmt.Sprintf("<span>Instance</span> %s/%s/%d", r.DeploymentName, r.JobName, *r.JobIndex))
	}

	if len(r.DeploymentName) > 0 {
		descPieces = append(descPieces, "<span>Deployment</span> "+r.DeploymentName)
	}

	if len(r.JobNameMatch) > 0 {
		descPieces = append(descPieces, "<span>Job match</span> "+r.JobNameMatch)
	}

	if len(r.JobName) > 0 {
		descPieces = append(descPieces, "<span>Job</span> "+r.JobName)
	}

	return template.HTML(strings.Join(descPieces, " "))
}

func (j Job) SelectedIndices(max int) ([]int, error) {
	if len(j.Indices) > 0 {
		var indices []int

		for _, index := range j.Indices {
			if index < max {
				indices = append(indices, index)
			}
		}

		return indices, nil
	}

	if len(j.Limit) > 0 {
		pieces := strings.Split(j.Limit, "-")

		if len(pieces) == 0 {
			return nil, bosherr.Errorf("Expected at least one integer in the limit '%s'", j.Limit)
		}

		var piecesN []int

		for _, piece := range pieces {
			pieceN, err := strconv.Atoi(strings.TrimSuffix(piece, "%"))
			if err != nil {
				return nil, bosherr.Errorf("Cannot convert '%s' to integer", piece)
			}

			piecesN = append(piecesN, pieceN)
		}

		hasPer := strings.HasSuffix(pieces[len(pieces)-1], "%")

		n := 0

		switch {
		case len(piecesN) == 1:
			n = piecesN[0]
		case len(piecesN) == 2:
			n = piecesN[0] + rand.Intn(piecesN[1]-piecesN[0])
		default:
			return nil, bosherr.Errorf("Expected at most two integers in the limit '%s'", j.Limit)
		}

		return rand.Perm(max)[0:numOrPercent(n, max, hasPer)], nil
	}

	return nil, bosherr.Errorf("Expected indices or limit specified")
}

func numOrPercent(n, max int, nIsPercent bool) int {
	if nIsPercent {
		return int(math.Ceil(float64(n) / 100.0 * float64(max)))
	}

	return n
}
