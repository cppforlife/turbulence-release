# API Go Client

See [turbulence-example-test/kill_test.go](./../src/github.com/cppforlife/turbulence-example-test/kill_test.go) and [turbulence-example-test/stress_test.go](./../src/github.com/cppforlife/turbulence-example-test/stress_test.go) for an example Ginkgo test that uses Go client.

## `kill_test.go` example

Initialize Turbulence client from the environment. 

```go
  var (
    client tubclient.Turbulence
  )

  BeforeEach(func() {
    logger := boshlog.NewLogger(boshlog.LevelNone)
    config := tubclient.NewConfigFromEnv()
    client = tubclient.NewFactory(logger).New(config)
  })
```

It currently expects following set of env variables like that:

```bash
export TURBULENCE_HOST=10.244.0.34
export TURBULENCE_PORT=8080
export TURBULENCE_USERNAME=turbulence
export TURBULENCE_PASSWORD=$(bosh int creds.yml --path /turbulence_api_password)
export TURBULENCE_CA_CERT=$(bosh int creds.yml --path /turbulence_api_cert/ca)
```

Initialize a single incident to kill VMs. Selector is set to choose `dummy` deployment with all instances that are part of `z1` AZ.

```go
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
```

Submit incident to the API server and Wait until request is fulfilled.

```go
  inc := client.CreateIncident(req)
  inc.Wait()
```

Assert that there were no errors while executing incident task.

```go
  Expect(inc.HasTaskErrors()).To(BeFalse())
```

`kill_test.go` asserts that all affected instances were in AZ `z1`; however, in real tests one would want to assert on some general system behaviour instead of just on Turbulence's behaviour. You should be able to pull necessary instance information if querying BOSH is necessary.

```go
  tasks := inc.TasksOfType(tubtasks.KillOptions{})
  Expect(tasks).To(HaveLen(4))

  for _, t := range tasks {
    Expect(t.Instance().AZ).To(Equal("z1"))
  }
```
