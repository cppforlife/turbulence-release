package incident

type Repo interface {
	ListAll() ([]Incident, error)
	Create(IncidentReq) (Incident, error)
	Read(string) (Incident, error)
}

type RepoNotifier interface {
	IncidentWasCreated(Incident)
}
