package example_test

import (
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"

  boshlog "github.com/cloudfoundry/bosh-utils/logger"
  tubclient "github.com/cppforlife/turbulence/client"
  tubinc "github.com/cppforlife/turbulence/incident"
  tubtasks "github.com/cppforlife/turbulence/tasks"
  tubsel "github.com/cppforlife/turbulence/incident/selector"
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
    req := tubinc.IncidentReq{
      Tasks: tubtasks.TaskOptionsSlice{
        tubtasks.KillOptions{},
      },

      Selector: tubsel.Req{
        Deployment: &tubsel.NameReq{Name: "dummy"},

        AZ: &tubsel.NameReq{
          Name: "z1",
          Limit: tubsel.MustNewLimitFromString("100%"),
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
