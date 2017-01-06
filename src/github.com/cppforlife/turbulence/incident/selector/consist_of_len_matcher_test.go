package selector_test

import (
	"errors"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func ConsistOfLen(length int, expected []string) types.GomegaMatcher {
	return &ConsistOfLenMatcher{Length: length, Expected: expected}
}

type ConsistOfLenMatcher struct {
	Length   int
	Expected []string
}

func (m ConsistOfLenMatcher) Match(actual interface{}) (bool, error) {
	actualStrings := actual.([]string)

	Expect(actualStrings).To(HaveLen(m.Length))

	err := m.checkUniq(actualStrings)
	if err != nil {
		return false, err
	}

	err = m.checkConsists(actualStrings)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m ConsistOfLenMatcher) checkUniq(actual []string) error {
	seenInts := []string{}

	for _, i := range actual {
		for _, j := range seenInts {
			if i == j {
				return errors.New(format.Message(actual, "to not contain duplicate", i))
			}
		}

		seenInts = append(seenInts, i)
	}

	return nil
}

func (m ConsistOfLenMatcher) checkConsists(actual []string) error {
	for _, i := range actual {
		var foundInExpected bool

		for _, j := range m.Expected {
			if i == j {
				foundInExpected = true
			}
		}

		if !foundInExpected {
			return errors.New(format.Message(actual, "to not contain", i))
		}
	}

	return nil
}

func (m ConsistOfLenMatcher) FailureMessage(_ interface{}) string { return "Not implemented" }

func (m ConsistOfLenMatcher) NegatedFailureMessage(_ interface{}) string { return "Not implemeted" }
