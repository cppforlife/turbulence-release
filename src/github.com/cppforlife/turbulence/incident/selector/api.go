package selector

type Req struct {
	IncludeMissing bool `json:",omitempty"`

	AZ         *NameReq `json:",omitempty"`
	Deployment *NameReq `json:",omitempty"`
	Group      *NameReq `json:",omitempty"`
	ID         *IDReq   `json:",omitempty"`
}

type NameReq struct {
	Name  string
	Limit Limit `json:",omitempty"`
}

type IDReq struct {
	Values []string `json:",omitempty"`
	Limit  Limit    `json:",omitempty"`
}

// todo State
// todo ProcessState
// todo PersistentDisk
// todo Bootstrap

func (a Req) AsSelector() Selector {
	selectors := []Selector{}

	// By default we avoid running any tasks against instances without VMs
	if !a.IncludeMissing {
		f := func(i Instance) (bool, error) { return i.HasVM(), nil }
		selectors = append(selectors, ByFilter{f})
	}

	if a.AZ != nil {
		f := func(i Instance) string { return i.AZ() }
		selectors = append(selectors, Generic{[]string{a.AZ.Name}, a.AZ.Limit, f})
	}

	if a.Deployment != nil {
		f := func(i Instance) string { return i.Deployment() }
		selectors = append(selectors, Generic{[]string{a.Deployment.Name}, a.Deployment.Limit, f})
	}

	if a.Group != nil {
		f := func(i Instance) string { return i.Group() }
		selectors = append(selectors, Generic{[]string{a.Group.Name}, a.Group.Limit, f})
	}

	if a.ID != nil {
		f := func(i Instance) string { return i.ID() }
		selectors = append(selectors, Generic{a.ID.Values, a.ID.Limit, f})
	}

	return Multiple{selectors}
}
