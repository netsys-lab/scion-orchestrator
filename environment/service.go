package environment

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/pkg/metrics"
)

var Services = map[string]*SystemService{}

type SystemService struct {
	Name       string
	BinaryPath string
	ConfigPath string
}

func GetControlServices() []*SystemService {
	services := make([]*SystemService, 0)
	for _, service := range Services {
		if strings.Contains(service.BinaryPath, "control") {
			services = append(services, service)
		}
	}
	return services
}

func GetBorderRouters() []*SystemService {
	services := make([]*SystemService, 0)
	for _, service := range Services {
		if strings.Contains(service.BinaryPath, "router") {
			services = append(services, service)
		}
	}
	return services
}

func UpdateHealthCheck() bool {
	allServicesRunning := true
	for _, service := range Services {
		if strings.Contains(service.BinaryPath, "control") && !service.IsRunning() {
			met := metrics.ServiceStatus{
				Status:  metrics.SERVICE_STATUS_ERROR,
				Message: "Service not running",
			}
			metrics.ASStatus.ControlServices[service.Name] = met
			allServicesRunning = false
		} else if strings.Contains(service.BinaryPath, "router") && !service.IsRunning() {
			met := metrics.ServiceStatus{
				Status:  metrics.SERVICE_STATUS_ERROR,
				Message: "Service not running",
			}
			metrics.ASStatus.BorderRouters[service.Name] = met
			allServicesRunning = false
		} else if strings.Contains(service.BinaryPath, "dispatcher") && !service.IsRunning() {
			met := metrics.ServiceStatus{
				Status:  metrics.SERVICE_STATUS_ERROR,
				Message: "Service not running",
			}
			metrics.ASStatus.Dispatcher = met
			allServicesRunning = false
		} else if strings.Contains(service.BinaryPath, "daemon") && !service.IsRunning() {
			met := metrics.ServiceStatus{
				Status:  metrics.SERVICE_STATUS_ERROR,
				Message: "Service not running",
			}
			metrics.ASStatus.Daemon = met
			allServicesRunning = false
		}

	}
	return allServicesRunning
}

func LoadServices(env *HostEnvironment, config *conf.SCIONConfig) error {

	binPath := "/usr/bin/"

	// TODO: Windows and MacOS Support!
	switch runtime.GOOS {
	case "linux":
		break
	}

	if config.Dispatcher != nil {
		service := &SystemService{
			Name:       config.Dispatcher.Name,
			BinaryPath: filepath.Join(binPath, "dispatcher"),
			ConfigPath: config.Dispatcher.ConfigFile,
		}

		Services[config.Dispatcher.Name] = service
		log.Println("[Env] Loaded Dispatcher Service")
	}

	service := &SystemService{
		Name:       config.Daemon.Name,
		BinaryPath: filepath.Join(binPath, "daemon"),
		ConfigPath: config.Daemon.ConfigFile,
	}

	Services[config.Daemon.Name] = service
	log.Println("[Env] Loaded Daemon Service")

	for _, service := range config.ControlServices {
		service := &SystemService{
			Name:       service.Name,
			BinaryPath: filepath.Join(binPath, "control"),
			ConfigPath: service.ConfigFile,
		}

		Services[service.Name] = service
		log.Println("[Env] Loaded Control Service: ", service.Name)
	}

	for _, service := range config.BorderRouters {
		service := &SystemService{
			Name:       service.Name,
			BinaryPath: filepath.Join(binPath, "router"),
			ConfigPath: service.ConfigFile,
		}

		Services[service.Name] = service
		log.Println("[Env] Loaded Border Router Service: ", service.Name)
	}

	service = &SystemService{
		Name:       "scion-as",
		BinaryPath: filepath.Join(binPath, "scion-as"),
		ConfigPath: filepath.Join(env.ConfigPath, "sciond.toml"),
	}

	Services[service.Name] = service
	log.Println("[Env] Loaded SCION-AS Service")
	return nil
}

func (s *SystemService) Install() error {
	return installService(s.Name, s.BinaryPath, s.ConfigPath)
}

func StartAllServices() error {
	errStr := ""

	// Some dependencies on dispatcher
	dispatcher, ok := Services["scion-dispatcher"]
	if ok && !dispatcher.IsRunning() {
		log.Println("[Env] service: ", dispatcher.Name, " is not running, starting...")
		err := dispatcher.Start()
		if err != nil {
			errStr += fmt.Sprintf("Error starting service %s: %v\n", dispatcher.Name, err)
		}
	}

	for _, service := range Services {
		if !service.IsRunning() && !strings.Contains(service.Name, "dispatcher") {
			log.Println("[Env] service: ", service.Name, " is not running, starting...")
			err := service.Start()
			if err != nil {
				errStr += fmt.Sprintf("Error starting service %s: %v\n", service.Name, err)
			}
			log.Println("[Env] Started service: ", service.Name)
		}
	}

	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func StopAllServices() error {
	errStr := ""
	for _, service := range Services {
		if service.IsRunning() {
			log.Println("[Env] service: ", service.Name, " is running, stopping...")
			err := service.Stop()
			if err != nil {
				errStr += fmt.Sprintf("Error stopping service %s: %v\n", service.Name, err)
			}
			log.Println("[Env] Stopped service: ", service.Name)
		}
	}

	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (s *SystemService) IsRunning() bool {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "is-active", "--quiet", s.Name)
		err := cmd.Run()
		fmt.Println("Is running: ", err == nil)
		if err != nil {
			return false
		}
		return true

		//bts, err := cmd.CombinedOutput()
		//if err != nil {
		//	return false
		//}

		//fmt.Println("Is running: ", string(bts))
		//return string(bts) == "active"
	case "darwin":
		return false
		// return installMacService(serviceName, binaryPath, configPath)
	case "windows":
		return false
		// return installWindowsService(serviceName, binaryPath, configPath)
	default:
		return false
	}
}

func (s *SystemService) Start() error {
	switch runtime.GOOS {
	case "linux":

		cmd := exec.Command("systemctl", "start", s.Name)
		err := cmd.Run()
		if err != nil {
			return err
		}

	case "darwin":
		return nil
		// return installMacService(serviceName, binaryPath, configPath)
	case "windows":
		return nil
		// return installWindowsService(serviceName, binaryPath, configPath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return nil
}

func (s *SystemService) ReStart() error {
	switch runtime.GOOS {
	case "linux":

		cmd := exec.Command("systemctl", "restart", s.Name)
		err := cmd.Run()
		if err != nil {
			return err
		}

	case "darwin":
		return nil
		// return installMacService(serviceName, binaryPath, configPath)
	case "windows":
		return nil
		// return installWindowsService(serviceName, binaryPath, configPath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return nil
}

func (s *SystemService) Stop() error {
	switch runtime.GOOS {
	case "linux":

		cmd := exec.Command("systemctl", "stop", s.Name)
		err := cmd.Run()
		if err != nil {
			return err
		}

	case "darwin":
		return nil
		// return installMacService(serviceName, binaryPath, configPath)
	case "windows":
		return nil
		// return installWindowsService(serviceName, binaryPath, configPath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return nil
}

func installService(serviceName, binaryPath, configPath string) error {
	switch runtime.GOOS {
	case "linux":
		return installLinuxService(serviceName, binaryPath, configPath)
	case "darwin":
		return installMacService(serviceName, binaryPath, configPath)
	case "windows":
		return installWindowsService(serviceName, binaryPath, configPath)
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func installLinuxService(serviceName, binaryPath, configPath string) error {
	unitFile := fmt.Sprintf(`[Unit]
Description=%s service
After=network.target

[Service]
ExecStart=%s --config %s
Restart=always
User=root
Group=root

[Install]
WantedBy=multi-user.target
`, serviceName, binaryPath, configPath)

	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	err := os.WriteFile(unitPath, []byte(unitFile), 0644)
	if err != nil {
		return err
	}
	fmt.Println("unitFile: ", unitPath)

	fmt.Println("Daemon reload: ", unitPath)
	// Reload systemd and enable the service
	cmd := exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("Enable: ", unitPath)
	cmd = exec.Command("systemctl", "enable", serviceName)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func installMacService(serviceName, binaryPath, configPath string) error {
	plistFile := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>--config</string>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`, serviceName, binaryPath, configPath)

	plistPath := fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName)
	err := os.WriteFile(plistPath, []byte(plistFile), 0644)
	if err != nil {
		return err
	}

	// Load the service
	cmd := exec.Command("launchctl", "load", plistPath)
	return cmd.Run()
}

func installWindowsService(serviceName, binaryPath, configPath string) error {
	cmd := exec.Command("sc", "create", serviceName, "binPath=", fmt.Sprintf(`"%s --config %s"`, binaryPath, configPath), "start=", "auto")
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("sc", "start", serviceName)
	return cmd.Run()
}

func init() {
	Services = make(map[string]*SystemService)
}

// Run the scion-as service and also install it as a service, then stop it when seeing that it runs as a service

/*func main() {
	serviceName := "example-binary"
	binaryPath := "/path/to/your/example-binary"
	configPath := "/path/to/your/sample.config"

	if err := installService(serviceName, binaryPath, configPath); err != nil {
		fmt.Printf("Error installing service: %v\n", err)
	} else {
		fmt.Println("Service installed successfully")
	}
}*/
