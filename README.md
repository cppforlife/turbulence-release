# Turbulence

Turbulence release is used for injecting different failure scenarios into a BOSH deployed system. Currently the following scenarios are supported:

- VM termination on BOSH supported IaaSes
- impose CPU/RAM/IO load
- network partitioning
- packet loss and delay

Release contains two jobs: `turbulence_api` and `turbulence_agent`.

API job is a server that provides management UI and accepts API requests to schedule and execute failure scenarios.

Agent job is a daemon that periodically retrieves instructions from the API server. It should be placed onto participating VMs.

Next steps:

- [Configuration doc](docs/config.md) on how to configure API server and agents
- [API doc](docs/api.md) on how to use Turbulence
- [API client doc](docs/client.md) on how to use Turbulence GO client
- [Development doc](docs/dev.md) on how to contribute

--
![](docs/home.png)
