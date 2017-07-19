package monit

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	serviceDescription = "vcap"
	listAllJobsScript  = `(get-wmiobject win32_service -filter "description='` + serviceDescription + `'") | ForEach{ $_.Name, $_.ProcessId }`
)

func (c httpClient) Services() ([]Service, error) {
	c.logger.Info("monit.Windows.httpClient", "Services: preparing to list services")

	cmd := exec.Command("Powershell.exe", "-Command", listAllJobsScript)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("monit.Services Error: %s Stderr: %s", err, stderr.String())
	}

	var svcs []Service
	lines := strings.Split(stdout.String(), "\n")

	for i := 0; i < len(lines)-1; i += 2 {
		name := strings.TrimSpace(lines[i])
		pid, err := strconv.Atoi(strings.TrimSpace(lines[i+1]))
		if name != "" && err == nil && pid > 0 {
			svcs = append(svcs, Service{Name: name, PID: pid})
		}
	}

	c.logger.Info("monit.Windows.httpClient", "Services: found %d services: %+v", len(svcs), svcs)
	return svcs, nil
}
