# Development

Source code is located in `src/github.com/cppforlife/turbulence`. Use `manifests/api.yml` and `manifests/agent.yml` to deploy API server and agents to BOSH Lite.

## Dependencies

Run `./update-deps` to update `github.com/cppforlife/turbulence` package dependencies. `deps.txt` will be updated with Git SHAs for each dependency.

## Planned tasks

- lock up whole machine
- remount disk as readonly
- corrupt disks
- pause a process
- restrict X% bandw

https://www.kernel.org/doc/Documentation/sysrq.txt might be useful...
