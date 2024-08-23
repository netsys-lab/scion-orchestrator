package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
)

func runInstall(env *environment.HostEnvironment, config *conf.SCIONConfig) error {

	// TODO: Proper error handling, do not fatal in here...
	// TODO: Mcast Bootstrapping, and all the other things too
	err := os.MkdirAll(env.ConfigPath, 0777)
	if err != nil {
		return err
	}

	log.Println("[Main] Installing files to ", env.BasePath)
	err = env.Install()
	if err != nil {
		return err
	}

	binPath := "/usr/bin/"

	// TODO: Windows and MacOS Support!
	switch runtime.GOOS {
	case "linux":
		break
	}

	if config.Dispatcher != nil {
		log.Println("[Install] Installing Dispatcher Service...")
		service := &environment.SystemService{
			Name:       config.Dispatcher.Name,
			BinaryPath: filepath.Join(binPath, "dispatcher"),
			ConfigPath: config.Dispatcher.ConfigFile,
		}

		err := service.Install()
		if err != nil {
			return err
		}
		environment.Services[config.Dispatcher.Name] = service
		log.Println("[Install] Installed Dispatcher Service")
	}

	log.Println("[Install] Installing Daemon Service...")
	service := &environment.SystemService{
		Name:       config.Daemon.Name,
		BinaryPath: filepath.Join(binPath, "daemon"),
		ConfigPath: config.Daemon.ConfigFile,
	}

	err = service.Install()
	if err != nil {
		return err
	}
	environment.Services[config.Daemon.Name] = service
	log.Println("[Install] Installed Daemon Service")

	for _, service := range config.ControlServices {
		log.Println("[Install] Installing Control Service: ", service.Name)
		service := &environment.SystemService{
			Name:       service.Name,
			BinaryPath: filepath.Join(binPath, "control"),
			ConfigPath: service.ConfigFile,
		}

		err := service.Install()
		if err != nil {
			return err
		}
		environment.Services[service.Name] = service
		log.Println("[Install] Installed Control Service: ", service.Name)
	}

	for _, service := range config.BorderRouters {
		log.Println("[Install] Installing Border Router Service: ", service.Name)
		service := &environment.SystemService{
			Name:       service.Name,
			BinaryPath: filepath.Join(binPath, "router"),
			ConfigPath: service.ConfigFile,
		}

		err := service.Install()
		if err != nil {
			return err
		}
		environment.Services[service.Name] = service
		log.Println("[Install] Installed Border Router Service: ", service.Name)
	}

	log.Println("[Install] Installing SCION-AS Service")
	service = &environment.SystemService{
		Name:       "scion-as",
		BinaryPath: filepath.Join(binPath, "scion-as"),
		ConfigPath: filepath.Join(env.ConfigPath, "sciond.toml"),
	}

	err = service.Install()
	if err != nil {
		return err
	}
	log.Println("[Install] SCION-AS Service installed")

	/*err = service.Start()
	if err != nil {
		return err
	}
	log.Println("[Main] Service started")

	for {
		if service.IsRunning() {
			break
		}
		time.Sleep(3 * time.Second)
	}
	log.Println("[Main] Service is running, closing this one now...")
	*/
	return nil
}
