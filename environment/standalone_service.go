package environment

import (
	"os"
	"os/exec"
	"strings"

	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
)

type StandaloneService struct {
	Name       string
	BinaryPath string
	ConfigPath string
	Logfile    string
}

// filepath.Join("bin", "control")
func (s *StandaloneService) Run() error {
	cmd := exec.Command(s.BinaryPath, "--config", s.ConfigPath)

	// Open s.Logfile for writing
	logfile, err := os.OpenFile(s.Logfile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	s.UpdateMetrics()

	cmd.Stderr = logfile
	cmd.Stdout = logfile
	processes = append(processes, cmd)
	return cmd.Run()
}

func (s *StandaloneService) UpdateMetrics() {
	lowerName := strings.ToLower(s.Name)
	if strings.Contains(lowerName, "daemon") {
		metrics.ASStatus.Daemon.Status = metrics.SERVICE_STATUS_RUNNING
	} else if strings.Contains(lowerName, "dispatcher") {
		metrics.ASStatus.Dispatcher.Status = metrics.SERVICE_STATUS_RUNNING
	} else if strings.Contains(lowerName, "control") {
		metrics.ASStatus.ControlServices[s.Name] = metrics.ServiceStatus{
			Status: metrics.SERVICE_STATUS_RUNNING,
		}
	} else if strings.Contains(lowerName, "router") {
		metrics.ASStatus.BorderRouters[s.Name] = metrics.ServiceStatus{
			Status: metrics.SERVICE_STATUS_RUNNING,
		}
	}
}
