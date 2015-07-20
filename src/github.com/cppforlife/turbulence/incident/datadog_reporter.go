package incident

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	datadog "github.com/zorkian/go-datadog-api"
)

type DatadogConfig struct {
	APIKey string
	AppKey string
}

type DatadogReporter struct {
	client *datadog.Client

	logTag string
	logger boshlog.Logger
}

func NewDatadogReporter(apiKey, appKey string, logger boshlog.Logger) DatadogReporter {
	return DatadogReporter{
		client: datadog.NewClient(apiKey, appKey),

		logTag: "DatadogReporter",
		logger: logger,
	}
}

func (r DatadogReporter) ReportIncidentExecutionStart(i Incident) {
	text, err := i.ShortDescription()
	if err != nil {
		text = fmt.Sprintf("Failed to generate incident description: %#v", i)
		r.logger.Error(r.logTag, text)
	}

	event := &datadog.Event{
		Title: r.incidentTitle("Started", i),
		Text:  text,
		Time:  int(i.ExecutionStartedAt.Unix()),

		Priority:  "normal",
		AlertType: "info",

		Host:        "turbulence-api",
		Aggregation: "",
		SourceType:  "turbulence-api",

		Tags:     []string{"incident:" + i.ID},
		Resource: "",
	}

	event, err = r.client.PostEvent(event)
	if err != nil {
		r.logger.Error(r.logTag, "Failed to send incident execution start event: %s event=%#v", err.Error(), event)
	} else {
		r.logger.Debug(r.logTag, "Posted incident '%s' execution start datadog event '%d'", i.ID, event.Id)
	}
}

func (r DatadogReporter) ReportIncidentExecutionCompletion(i Incident) {
	text := ""
	alertType := "info"

	incidentErr := i.Events.FirstError()
	if incidentErr != nil {
		text = fmt.Sprintf("Error: %s", incidentErr.Error())
		alertType = "error"
	}

	event := &datadog.Event{
		Title: r.incidentTitle("Completed", i),
		Text:  text,
		Time:  int(i.ExecutionCompletedAt.Unix()),

		Priority:  "normal",
		AlertType: alertType,

		Host:        "turbulence-api",
		Aggregation: "",
		SourceType:  "turbulence-api",

		Tags:     []string{"incident:" + i.ID},
		Resource: "",
	}

	event, err := r.client.PostEvent(event)
	if err != nil {
		r.logger.Error(r.logTag, "Failed to send incident execution completion event: %s event=%#v", err.Error(), event)
	} else {
		r.logger.Debug(r.logTag, "Posted incident '%s' execution completion datadog event '%d'", i.ID, event.Id)
	}
}

func (r DatadogReporter) ReportEventExecutionStart(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	event := &datadog.Event{
		Title: r.eventTitle("Started", e),
		Text:  "",
		Time:  int(e.ExecutionStartedAt.Unix()),

		Priority:  "normal",
		AlertType: "info",

		Host:        "turbulence-api",
		Aggregation: "",
		SourceType:  "turbulence-api",

		Tags:     r.eventTags(incidentID, e),
		Resource: "",
	}

	event, err := r.client.PostEvent(event)
	if err != nil {
		r.logger.Error(r.logTag, "Failed to send event execution completion event: %s event=%#v", err.Error(), event)
	} else {
		r.logger.Debug(r.logTag, "Posted event '%s' execution start datadog event '%d'", e.ID, event.Id)
	}
}

func (r DatadogReporter) ReportEventExecutionCompletion(incidentID string, e Event) {
	if !e.IsAction() {
		return
	}

	text := ""
	alertType := "info"

	if e.Error != nil {
		text = fmt.Sprintf("Error: %s", e.Error.Error())
		alertType = "error"
	}

	event := &datadog.Event{
		Title: r.eventTitle("Completed", e),
		Text:  text,
		Time:  int(e.ExecutionCompletedAt.Unix()),

		Priority:  "normal",
		AlertType: alertType,

		Host:        "turbulence-api",
		Aggregation: "",
		SourceType:  "turbulence-api",

		Tags:     r.eventTags(incidentID, e),
		Resource: "",
	}

	event, err := r.client.PostEvent(event)
	if err != nil {
		r.logger.Error(r.logTag, "Failed to send event execution completion event: %s event=%#v", err.Error(), event)
	} else {
		r.logger.Debug(r.logTag, "Posted event '%s' execution completion datadog event '%d'", e.ID, event.Id)
	}
}

func (r DatadogReporter) incidentTitle(prefix string, i Incident) string {
	return fmt.Sprintf("%s incident '%s': %s",
		prefix, i.ID, strings.Join(i.TaskTypes(), ", "))
}

func (r DatadogReporter) eventTitle(prefix string, e Event) string {
	return fmt.Sprintf("%s event '%s': %s for %s/%s/%d",
		prefix, e.ID, e.Type, e.DeploymentName, e.JobName, *e.JobIndex)
}

func (r DatadogReporter) eventTags(incidentID string, e Event) []string {
	return []string{
		"incident:" + incidentID,
		"event:" + e.ID,
		"deployment:" + e.DeploymentName,
		"job:" + e.JobName,
		fmt.Sprintf("index:%d", *e.JobIndex),
		fmt.Sprintf("instance:%s/%s/%d", e.DeploymentName, e.JobName, *e.JobIndex),
	}
}

func (c DatadogConfig) Required() bool { return len(c.AppKey) > 0 }

func (c DatadogConfig) Validate() error {
	if !c.Required() {
		return nil
	}

	if len(c.APIKey) == 0 {
		return bosherr.Error("Missing 'APIKey'")
	}

	return nil
}
