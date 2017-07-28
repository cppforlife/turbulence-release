package example_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	tubclient "github.com/cppforlife/turbulence/client"
	tubinc "github.com/cppforlife/turbulence/incident"
	tubsel "github.com/cppforlife/turbulence/incident/selector"
	tubtasks "github.com/cppforlife/turbulence/tasks"
)

var _ = Describe("Kill", func() {
	var (
		client tubclient.Turbulence
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		config := tubclient.NewConfigFromEnv()
		client = tubclient.NewFactory(logger).New(config)
	})

	It("kills dummy deployment's z1", func() {
		req := tubinc.Request{
			Tasks: tubtasks.OptionsSlice{
				tubtasks.KillOptions{},
			},

			Selector: tubsel.Request{
				Deployment: &tubsel.NameRequest{Name: "dummy"},

				AZ: &tubsel.NameRequest{
					Name: "z1",
				},
			},
		}

		{ // Check that kill kills all z1 instances
			inc := client.CreateIncident(req)
			inc.Wait()

			Expect(inc.HasEventErrors()).To(BeFalse())

			events := inc.EventsOfType(tubtasks.KillOptions{})
			Expect(events).To(HaveLen(4))

			for _, ev := range events {
				Expect(ev.Instance.AZ).To(Equal("z1"))
			}
		}
	})
})
