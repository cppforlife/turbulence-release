package scheduledinc

type NotFoundError struct {
	ID string
}

type Repo interface {
	ListAll() ([]ScheduledIncident, error)
	Create(Request) (ScheduledIncident, error)
	Read(string) (ScheduledIncident, error)
	Delete(string) error
}

type RepoNotifier interface {
	ScheduledIncidentWasCreated(ScheduledIncident)
	ScheduledIncidentWasDeleted(ScheduledIncident)
}
