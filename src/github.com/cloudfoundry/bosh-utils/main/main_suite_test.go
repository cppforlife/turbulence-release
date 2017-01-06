package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"github.com/onsi/gomega/gexec"
)

var pathToBoshUtils string

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	BeforeSuite(func() {
		var err error
		pathToBoshUtils, err = gexec.Build("github.com/cloudfoundry/bosh-utils/main")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})
	RunSpecs(t, "Main Suite")
}
