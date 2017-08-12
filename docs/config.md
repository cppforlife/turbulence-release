# Configuration

## API server configuration

API server job is configured to serve over SSL (required) on two ports (`listen_port` and `agent_listen_port` properties) for operator and agent APIs.

Currently basic auth is used for UI and API access by an operator and agents, but we have plans to secure it via UAA integration (todo).

API server uses Director API to find all instances in all deployments. It also can issue delete VM API calls (equivalent to `bosh delete-vm VMCID` command) when Kill task is requested. It's recommend to configure API server with a didicated Director user so that it's easier to see its activity via events command (i.e. `bosh events --user turbulence`).

Director UAA integration is supported.

```
$ bosh -n -d turbulence deploy ./manifests/example.yml \
  -v turbulence_api_ip=10.244.0.34 \
  -v director_ip=192.168.50.6 \
  -v director_client=turbulence \
  -v director_client_secret=... \
  --var-file director_ssl.ca=/tmp/director-ca \
  --vars-store /tmp/turbulence.yml

$ bosh int /tmp/turbulence.yml --path /turbulence_api_password
...
```

Events emitted (found via `bosh events`):

- actions: `start` or `end`, object type: `turbulence-incident`, object name: `<incident id>`
- actions: `start` or `end`, object type: `turbulence-event`, object name: `<event id>`

## Agent configuration

Agent job is configured to communicate with the API server. Communication is done over SSL with basic auth.

```yaml
instance_groups:
- name: cell
  azs: [z1, z2]
  instances: 10
  jobs:
  - name: executor
    release: diego
  - name: turbulence_agent
    release: turbulence
    consumes:
      api: {from: api, deployment: turbulence}
  vm_type: default
  stemcell: default
  networks:
  - name: default
```

## Datadog configuration

API server can be configured to post events to Datadog for easier event correlation.

```
$ bosh -n -d turbulence deploy ./manifests/example.yml \
  -o ./manifests/datadog.yml \
  -v datadog_app_key=... \
  -v datadog_api_key=... \
  ...
```
