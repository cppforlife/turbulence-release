package incident_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStemsrepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "incident")
}
