package environment

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type SystemService struct {
	Name       string
	BinaryPath string
	ConfigPath string
}

func (s *SystemService) Install() error {
	return installService(s.Name, s.BinaryPath, s.ConfigPath)
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
