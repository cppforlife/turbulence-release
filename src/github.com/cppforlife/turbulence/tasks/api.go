package tasks

type StateRequest struct {
	Stop bool
}

type StateResponse struct {
	Stop bool
}

type ResultRequest struct {
	Error string
}
