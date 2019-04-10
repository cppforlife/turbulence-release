# API

## Incidents

API server can perform several types of tasks against a subset of instances in several deployments. Types of tasks and selection of instances is represented by an incident.

See 'Incident Types' section below for details on which task types are available.

Instance selection is based on `Deployments` (array of hashes; required) configuration. An incident may affect one or deployments from which one or more jobs can be selected. For each job selected, its instances are filtered via `Indices` (array of ints) or `Limit` (string) keys. Limit value may be one of the following:

- `10`: kill 10 *random* instances
- `5-10`: kill 5-10 *random* instances
- `10%`: kill random 10% of instances
- `5%-10%`: kill random 5% to 10% of instances

Endpoints:

- `POST /api/v1/incidents`
- `GET /api/v1/incidents`
- `GET /api/v1/incidents/:id`

```
$ curl -vvv -k -X POST https://user:pass@10.244.8.2:8080/api/v1/incidents -H 'Accept: application/json' -d@example.json
```

See [docs/kill.sh](kill.sh) for live example.

Create request:

```json
{
	"Tasks": [{
		"Type": "Stress",
		"Timeout": "10m",

		"NumCPUWorkers": 1
	}],

	"Selector": {
		"Deployment": {
			"Name": "cf"
		},
		"Group": {
			"Name": "postgres",
		},
		"ID": {
			"Limit": "30%"
		}
	}
}
```

Response:

```json
{
  "ID": "d77adc3b-1de4-4e12-4bee-b325adfbecbd",

  "Tasks": [ ... ],
  "Selector": { ... },

  "ExecutionStartedAt": "0001-01-01T00:00:00Z",
  "ExecutionCompletedAt": "",

  "Events": null
}
```

Available selector rules:

- AZ
  - set `Name` (string; optional)
  - set `Limit` (string; optional)

- Deployment
  - set `Name` (string; optional)
  - set `Limit` (string; optional)

- Group
  - set `Name` (string; optional)
  - set `Limit` (string; optional)

- ID
  - set `Values` (array of strings; optional)
  - set `Limit` (string; optional)

Limits default to 100%. Name defaults to '\*' and wildcard matches are supported.

```json
{
	"AZ": {
		"Name": "z1",
		"Limit": "5"
	},
	"Deployment": {
		"Name": "cf-*",
		"Limit": "5"
	},
	"Group": {
		"Name": "postgres",
		"Limit": "5"
	},
	"ID": {
		"Values": ["53c5ae69-4622-4103-9766-230adcf3baef"],
		"Limit": "5"
	}
}
```

See [docs/selector-examples.md](selector-examples.md) for additional options.

---
## Scheduled Incidents

API server can create new incidents based on a schedule. `Schedule` (string; required) can be specified in cron format or with one of the shorthands:

- `@yearly`
- `@monthly`
- `@weekly`
- `@daily`
- `@hourly`
- `@every X` where X is value accepted by the [golang's time.ParseDuration](http://golang.org/pkg/time/#ParseDuration)

`Incident` (hash; required) is specified in the exactly the same way as when creating a single incident.

Endpoints:

- `POST /api/v1/scheduled_incidents`
- `GET /api/v1/scheduled_incidents`
- `GET /api/v1/scheduled_incidents/:id`
- `DELETE /api/v1/scheduled_incidents/:id`

```
$ curl -vvv -k -X POST https://user:pass@10.244.8.2:8080/api/v1/scheduled_incidents -H 'Accept: application/json' -d@example.json
```

See [docs/kill-scheduled.sh](kill-scheduled.sh) for live example.

Create request:

```json
{
	"Schedule": "@every 1m",
	"Incident": { ... }
}
```

Response:

```json
{
  "ID": "bf43eed7-91c7-4983-5895-44b9a18a5461",

  "Schedule": "@every 1m",
  "Incident": { ... }
}
```

---
## Incident Tasks

Currently there are eight support task types that can be included in an incident. Some tasks require `Timeout` key to be set so that the task can complete.

### Noop

Does not do anything on selected instances. This may be used for testing selection logic or communication between agents and the API server.

Optionally specify:

- set `Stoppable` (bool) to true to control whether task will block until it's stopped via an API

Example:

```json
{
  "Type": "Noop"
}
```

### Kill

Deletes the VM associated with an instance. API server uses newer Director API that is equivalent to using `bosh delete-vm VMCID` command.

Example:

```json
{
	"Type": "Kill"
}
```

### Kill Process

Kill one or more processes on the VM associated with an instance.

One of the following configurations must be selected:

- set `ProcessName` (string) to a pattern used with `pkill`
- set `MonitoredProcessName` (string) to a name of one of the processes watched by Monit
- by default random monitored process is killed

Example:

```json
{
	"Type": "KillProcess",
	"MonitoredProcessName": "*worker*"
}
```

### Pause Process

Pause one or more process on the VM associated with an instance.

Configuration:

- set `ProcessName`(string) to a pattern used with `pkill`
- set `timeout` (string) to how long the process should remain paused. A valid timeout is required.

Example:

```json
{
	"Type": "PauseProcess",
	"ProcessName": "sshd",
	"Timeout": "10m" // Times may be suffixed with s,m,h,d,y
}
```

### Stress

Stresses different subsystems on the VM associated with an instance.

One or more of the following configurations must be selected:

- CPU
  - set `NumCPUWorkers` (int; required)

- IO
  - set `NumIOWorkers` (int; required)

- RAM
  - set `NumMemoryWorkers` (int; required)
  - set `MemoryWorkerBytes` (string; required). Must be suffixed with B,K,M,G.

- HDD
  - set `NumHDDWorkers` (int; required)
  - set `HDDWorkerBytes` (string; required). Must be suffixed with B,K,M,G.

Example:

```json
{
	"Type": "Stress",
	"Timeout": "10m", // Times may be suffixed with s,m,h,d,y

	"NumCPUWorkers": 1,

	"NumIOWorkers": 1,

	"NumMemoryWorkers": 1,
	"MemoryWorkerBytes": "10K"
}
```

### Firewall

Blocks incoming and outgoing traffic from the VM associated with an instance. Useful for simulating network partitions. By default BOSH Agent and SSH on the VM will continue to operate.

Currently iptables is used for dropping packets from INPUT and OUTPUT chains.

Optionally specify:

- set `BlockBOSHAgent` (bool) to true to block access to the BOSH Agent

Example:

```json
{
	"Type": "Firewall",
	"Timeout": "10m" // Times may be suffixed with ms,s,m,h
}
```

### TargetedBlocker

Drops incoming and or outgoing traffic from one or more VMs. It is able to target specific IPs and Ports to simulate the failure of specific services.

Currently iptables is used for dropping packets from INPUT and OUTPUT chains.

Target parameters:

- set `Direction` (string; required) to the direction of traffic to drop, can be either "INPUT", "OUTPUT", or "FORWARD". If you are targeting diego-cells, then you will probably want "FORWARD".
- set `SrcHost` (string) to either an IPv4 address such as "192.168.1.50" or with a mask such as "192.168.0.0/24", or to a domain name which will be resolved into (possibly multiple) IPs such as "example.com" using the dig command. If no host is specified, then all source hosts will be impacted.
- set `DstHost` (string) to either an IPv4 address such as "192.168.1.50" or with a mask such as "192.168.0.0/24", or to a domain name which will be resolved into (possibly multiple) IPs such as "example.com" using the dig command. If no host is specified, then all destination hosts will be impacted.
- set `Protocol` (string) to the protocol to drop traffic on, can be either "udp", "tcp", "icmp", or "all". Defaults to being unspecified.
- set `DstPorts` (string) to the destination port to drop. This can be either a single port such as "8080" or a range such as "1503:1520". If blank, all destination ports will be dropped.
- set `SrcPorts` (string) to the source ports to drop. This can be either a single port such as "8080" or a range such as "1503:1520". If blank, all source ports will be dropped.

*Note*: at least one of `SrcHost`, `DstHost`, `DstPorts`, or `SrcPorts` must be specified.

Example:

```json
{
	"Type": "TargetedBlocker",
	"Timeout": "10m", // Times may be suffixed with ms,s,m,h
	"Targets": [{
		"DstHost": "1.1.1.1",
		"Direction": "INPUT",
		"DstPorts": "53"
	},{
		"DstHost": "google.com",
		"Direction": "FORWARD",
		"Protocol": "tcp",
		"DstPorts": "80"
	}]
}
```


### Block DNS

Causes all outgoing DNS packets to be dropped.

Currently iptables is used for dropping packets going out on tcp or udp port 53.

Example:

```json
{
	"Type": "BlockDNS",
	"Timeout": "10m" // Times may be suffixed with ms,s,m,h
}
```

### Control Network

Controls network quality on the VM associated with an instance. Does not affect `lo0`.

Currently [tc](http://www.lartc.org/manpages/tc.txt) is used to control package delay and loss.

One or both of the following configurations must be selected:

- packet delay
  - set `Delay` (string; required). Must be suffixed with `ms`.
  - set `DelayVariation` (string; optional). Must be suffixed with `ms`. Default is `10ms`.
  - if `DelayVariation >= 0.5*Delay`, then packet reordering may occur.

- packet loss
  - set `Loss` (string; required). Must be suffixed with `%`.
  - set `LossCorrelation` (string; optional). Must be suffixed with `%`. Default is `75%`.
  
- packet duplication
  - set `Duplication` (string; required). Must be suffixed with `%`.
  
- packet corruption
  - set `Corruption` (string; required). Must be suffixed with `%`.
  
- packet reordering
  - set `Reorder` (string; required). Must be suffixed with `%`.
  - set `ReorderCorrelation` (string; optional). Must be suffixed with `%`. Default is `50%`.
  - if the `Delay` is less than the inter-packet arrival time, then no reordering will be observed.
  
- bandwidth limiting
  - set `Bandwidth` (string; required). Must be suffixed with one of `kbps`, `mbps` or `gbps`.
  - bandwidth limiting must be used without any other effects.

In addition it is possible to apply a destination filter:
  - set `Targets` (array, optional). Must include either `DstHost` or `DstPort`

Example:

```json
{
	"Type": "ControlNet",
	"Timeout": "10m", // Times may be suffixed with ms,s,m,h

	"Delay": "50ms",
	
	"Targets": [
	  {
		"DstHost": "1.2.3.4",
		"DstHost": "443"
	  }
	]
}
```

### Fill Disk

Fill specific disk location on the VM associated with an instance.

One of the following configurations must be selected:

- set `Persistent` (bool) to fill up /var/vcap/store
- set `Ephemeral` (bool) to fill up /var/vcap/data
- set `Temporary` (bool) to fill up /tmp
- by default uses root disk
- if multiple are selected, the first one in the above order will be used.

Example:

```json
{
	"Type": "FillDisk",
	"Timeout": "10m", // Times may be suffixed with ms,s,m,h
	"Persistent": true
}
```

### Shutdown

Shuts down the VM associated with an instance.

One of the following configurations must be selected:

- set `Crash` (bool) to crash the system
- set `Reboot` (bool) to nicely reboot the system
- set `Sysrq` (string) to specify custom [system request](https://www.kernel.org/doc/Documentation/sysrq.txt)
- by default system will be nicely powered off

In addition you can set `Force` (bool) to forcefully reboot or power off.

Example:

```json
{
	"Type": "Shutdown",
	"Crash": true
}
```
