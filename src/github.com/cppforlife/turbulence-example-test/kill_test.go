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

		config, err := tubclient.NewConfigFromEnv()
		Expect(err).ToNot(HaveOccurred())

		client, err = tubclient.NewFactory(logger).New(config)
		Expect(err).ToNot(HaveOccurred())
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

		inc, err := client.CreateIncident(req)
		Expect(err).ToNot(HaveOccurred())

		err = inc.Wait()
		Expect(err).ToNot(HaveOccurred())

		Expect(inc.HasEventErrors()).To(BeFalse())

		events := inc.EventsOfType(tubtasks.KillOptions{})
		Expect(events).To(HaveLen(4))

		for _, ev := range events {
			Expect(ev.Instance.AZ).To(Equal("z1"))
		}
	})
})
