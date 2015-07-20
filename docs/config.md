# Configuration

## API server configuration

API server job is configured to serve over SSL. Basic auth is used for UI and API access by an operator and agents.

To support VM termination scenario, API server can be optionally collocated a CPI release.

```yaml
jobs:
- name: api
  templates:
  - {name: turbulence_api, release: turbulence}
  - {name: cpi, release: bosh-warden-cpi}
  instances: 1
  resource_pool: default
  networks:
  - {name: default, static_ips: [10.244.8.2]}

properties:
  turbulence_api:
    password: turbulence-password
    certificate: |
      -----BEGIN CERTIFICATE-----
      MIID...snip...
    private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIIE...snip...

    director:
      host: 192.168.50.4
      username: admin
      password: admin
      ca_cert: |
        -----BEGIN CERTIFICATE-----
        MIIDt...snip...

  cpi:
    warden:
      connect_network: tcp
      connect_address: 10.254.50.4:7777
    agent:
      mbus: nats://nats:nats-password@10.254.50.4:4222
      blobstore:
        provider: dav
        options:
          endpoint: http://10.254.50.4:25251
          user: agent
          password: agent-password
```

## Agent configuration

Agent job is configured to communicate with the API server. Communication is done over SSL with basic auth.

```yaml
jobs:
- name: dea_next_z1
  templates:
  - {name: dea_next, release: cf}
  - {name: turbulence_agent, release: turbulence}
  instances: 10
  resource_pool: default
  networks:
  - name: default

properties:
  turbulence_agent:
    api:
      host: 10.244.8.2
      password: turbulence-password
      ca_cert: |
        -----BEGIN CERTIFICATE-----
        MIIDt...snip...
```

## Datadog configuration

API server can be configured to post events to Datadog for easier event correlation.

```yaml
properties:
  turbulence_api:
  	datadog:
      app_key: 280b13972ebce1a6ff01b38970b6463fa18873c1
      api_key: f41bd13281ce18641312b496bc370184

    ...snip...
```
