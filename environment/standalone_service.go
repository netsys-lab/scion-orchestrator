package environment

import (
	"os"
	"os/exec"
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

	cmd.Stderr = logfile
	cmd.Stdout = logfile
	processes = append(processes, cmd)
	return cmd.Run()
}
