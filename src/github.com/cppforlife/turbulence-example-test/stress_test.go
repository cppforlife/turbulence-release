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

var _ = Describe("Stress", func() {
	var (
		client tubclient.Turbulence
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		config := tubclient.NewConfigFromEnv()
		client = tubclient.NewFactory(logger).New(config)
	})

	It("stresses dummy deployment's z2", func() {
		req := tubinc.Request{
			Tasks: tubtasks.OptionsSlice{
				tubtasks.StressOptions{
					Timeout: "30s",
				},
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

		{ // Check that execution of an invalid task returns an error
			inc := client.CreateIncident(req)
			inc.Wait()

			Expect(inc.HasTaskErrors()).To(BeTrue())

			tasks := inc.TasksOfType(tubtasks.StressOptions{})
			Expect(tasks).To(HaveLen(1))
			Expect(tasks[0].Error()).To(ContainSubstring("Task execution: Must specify at least 1 type of worker"))
		}

		req.Tasks = tubtasks.OptionsSlice{
			tubtasks.StressOptions{
				Timeout:       "30s",
				NumCPUWorkers: 1,
			},
		}

		{ // Check that stress can succeed
			inc := client.CreateIncident(req)
			inc.Wait()

			Expect(inc.HasTaskErrors()).To(BeFalse())

			tasks := inc.TasksOfType(tubtasks.StressOptions{})
			Expect(tasks).To(HaveLen(1))

			duration := inc.ExecutionCompletedAt().Sub(inc.ExecutionStartedAt())
			Expect(duration).To(BeNumerically(">=", 30*time.Second))
		}

		req.Tasks = tubtasks.OptionsSlice{
			tubtasks.StressOptions{
				NumCPUWorkers: 1,
			},
		}

		{ // Check that stress can be stopped successfully
			inc := client.CreateIncident(req)

			time.Sleep(10 * time.Second)

			tasks := inc.TasksOfType(tubtasks.StressOptions{})
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
