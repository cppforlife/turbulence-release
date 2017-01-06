package selector

import (
	"path/filepath"
)

type Multiple struct {
	Selectors []Selector
}

func (m Multiple) Select(instances []Instance) ([]Instance, error) {
	for _, sel := range m.Selectors {
		var err error

		instances, err = sel.Select(instances)
		if err != nil {
			return nil, err
		}
	}

	return instances, nil
}

type Generic struct {
	Names []string
	Limit Limit
	Func  func(Instance) string
}

func (g Generic) Select(instances []Instance) ([]Instance, error) {
	var err error

	if len(g.Names) > 0 {
		instances, err = NewByNames(g.Names, g.Func).Select(instances)
		if err != nil {
			return nil, err
		}
	}

	groups, err := g.Limit.Limit(GroupBy{g.Func}.Groups(instances))
	if err != nil {
		return nil, err
	}

	return NewByNames(groups, g.Func).Select(instances)
}

func NewByNames(names []string, f func(Instance) string) ByFilter {
	byNames := func(inst Instance) (bool, error) {
		for _, name := range names {
			matched, err := filepath.Match(name, f(inst)) // todo better matching
			if matched || err != nil {
				return matched, err
			}
		}
		return false, nil
	}

	return ByFilter{byNames}
}

type ByFilter struct {
	Func func(Instance) (bool, error)
}

func (s ByFilter) Select(instances []Instance) ([]Instance, error) {
	var matchedInst []Instance

	for _, inst := range instances {
		matched, err := s.Func(inst)
		if err != nil {
			return nil, err
		}

		if matched {
			matchedInst = append(matchedInst, inst)
		}
	}

	return matchedInst, nil
}

type GroupBy struct {
	Func func(Instance) string
}

func (g GroupBy) Groups(instances []Instance) []string {
	groupsMap := map[string]struct{}{}
	groups := []string{}

	for _, inst := range instances {
		groupsMap[g.Func(inst)] = struct{}{}
	}

	for g, _ := range groupsMap {
		groups = append(groups, g)
	}

	return groups
}
