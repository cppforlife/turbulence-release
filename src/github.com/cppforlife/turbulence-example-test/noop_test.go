package example_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	tubclient "github.com/cppforlife/turbulence/client"
	tubinc "github.com/cppforlife/turbulence/incident"
	tubsel "github.com/cppforlife/turbulence/incident/selector"
	tubtasks "github.com/cppforlife/turbulence/tasks"
)

var _ = Describe("Noop", func() {
	var (
		client tubclient.Turbulence
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		config := tubclient.NewConfigFromEnv()
		client = tubclient.NewFactory(logger).New(config)
	})

	It("noops dummy deployment's z2", func() {
		req := tubinc.Request{
			Tasks: tubtasks.OptionsSlice{
				tubtasks.NoopOptions{},
			},

			Selector: tubsel.Request{
				Deployment: &tubsel.NameRequest{Name: "dummy"},

				AZ: &tubsel.NameRequest{
					Name: "z2",
				},

				ID: &tubsel.IDRequest{
					Limit: tubsel.MustNewLimitFromString("50%"),
				},
			},
		}

		{ // Check that noop can run and finish
			inc := client.CreateIncident(req)
			inc.Wait()

			Expect(inc.HasTaskErrors()).To(BeFalse())

			tasks := inc.TasksOfType(tubtasks.NoopOptions{})
			Expect(tasks).To(HaveLen(1))

			duration := inc.ExecutionCompletedAt().Sub(inc.ExecutionStartedAt())
			Expect(duration).To(BeNumerically("<=", 5*time.Second))
		}

		req.Tasks = tubtasks.OptionsSlice{
			tubtasks.NoopOptions{
				Stoppable: true,
			},
		}

		{ // Check that we can stop it
			inc := client.CreateIncident(req)

			time.Sleep(10 * time.Second)

			tasks := inc.TasksOfType(tubtasks.NoopOptions{})
			Expect(tasks).To(HaveLen(1))

			tasks[0].Stop()
			inc.Wait()

			Expect(inc.HasTaskErrors()).To(BeFalse())

			duration := inc.ExecutionCompletedAt().Sub(inc.ExecutionStartedAt())
			Expect(duration).To(BeNumerically(">=", 10*time.Second))
			Expect(duration).To(BeNumerically("<=", 30*time.Second))
		}
	})
})
