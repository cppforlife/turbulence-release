package incident

type Repo interface {
	ListAll() ([]Incident, error)
	Create(Request) (Incident, error)
	Read(string) (Incident, error)
}

type RepoNotifier interface {
	IncidentWasCreated(Incident)
}
