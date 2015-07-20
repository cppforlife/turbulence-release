package client_test

import (
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	"testing"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Davcli Client Suite")
}
