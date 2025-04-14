package environment

import (
	"os"
	"os/exec"
	"strings"

	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
)

var StandaloneServices = map[string]*StandaloneService{}

type StandaloneService struct {
	Name       string
	BinaryPath string
	ConfigPath string
	Logfile    string
	stopChan   chan struct{}
	doneChan   chan error
}

// TODO: Restarting does not work here...
func (s *StandaloneService) Run() error {
	// Initialize channels if not already done
	if s.stopChan == nil {
		s.stopChan = make(chan struct{})
	}
	if s.doneChan == nil {
		s.doneChan = make(chan error)
	}

	go func() {
		for {
			cmd := exec.Command(s.BinaryPath, "--config", s.ConfigPath)

			// Open s.Logfile for writing
			logfile, err := os.OpenFile(s.Logfile, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				s.doneChan <- err
				return
			}
			defer logfile.Close()

			s.UpdateMetrics()

			cmd.Stderr = logfile
			cmd.Stdout = logfile

			// Start the process
			if err := cmd.Start(); err != nil {
				s.doneChan <- err
				return
			}

			// Wait for the process to finish or stop signal
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-s.stopChan:
				// Stop signal received, kill the process
				_ = cmd.Process.Kill()
				s.doneChan <- nil
				return
			case err := <-done:
				// Process finished, send the result
				s.doneChan <- err
				return
			}
		}
	}()

	// Block until the process finishes or an error occurs
	return <-s.doneChan
}

func (s *StandaloneService) Restart() error {
	// Stop the service if it's running
	if err := s.Stop(); err != nil {
		return err
	}

	// Start the service again
	return s.Run()
}

func (s *StandaloneService) Stop() error {
	// Signal the stop channel to terminate the process
	if s.stopChan != nil {
		close(s.stopChan)
		s.stopChan = nil
	}

	// Wait for the process to finish
	if s.doneChan != nil {
		<-s.doneChan
	}
	return nil
}

func (s *StandaloneService) UpdateMetrics() {
	lowerName := strings.ToLower(s.Name)
	if strings.Contains(lowerName, "daemon") {
		metrics.Status.Daemon.Status = metrics.SERVICE_STATUS_RUNNING
	} else if strings.Contains(lowerName, "dispatcher") {
		metrics.Status.Dispatcher.Status = metrics.SERVICE_STATUS_RUNNING
	} else if strings.Contains(lowerName, "control") {
		metrics.Status.ControlServices[s.Name] = metrics.ServiceStatus{
			Status: metrics.SERVICE_STATUS_RUNNING,
			Id:     s.Name,
		}
	} else if strings.Contains(lowerName, "router") {
		metrics.Status.BorderRouters[s.Name] = metrics.ServiceStatus{
			Status: metrics.SERVICE_STATUS_RUNNING,
			Id:     s.Name,
		}
	}
}
