package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/environment"
	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
)

// TODO: Might be deprecated for api/v1/install
func runInstall(env *environment.HostEnvironment, config *conf.SCIONConfig, asConfig *conf.Config) error {

	// TODO: Binary copy does not work when services are running

	// TODO: Proper error handling, do not fatal in here...
	// TODO: Mcast Bootstrapping, and all the other things too
	err := os.MkdirAll(env.ConfigPath, 0777)
	if err != nil {
		return err
	}

	err = environment.LoadServices(env, config, asConfig)
	if err != nil {
		return err
	}

	log.Println("[Install] Stopping all services...")
	err = environment.StopAllServices()
	if err != nil {
		return err
	}

	// Systemd might need a moment to stop the services
	time.Sleep(5 * time.Second)

	log.Println("[Install] Installing files to ", env.BasePath)
	err = env.Install(nil)
	if err != nil {
		return err
	}

	if config.Dispatcher != nil && !asConfig.ServiceConfig.DisableDispatcher {
		log.Println("[Install] Installing Dispatcher Service...")
		service, ok := environment.Services[config.Dispatcher.Name]
		if !ok {
			log.Println("[Install] Dispatcher Service not found in environment, name mismatch...")
			return fmt.Errorf("dispatcher Service not found in environment, name mismatch")
		}

		err := service.Install()
		if err != nil {
			return err
		}
		log.Println("[Install] Installed Dispatcher Service")
	}

	if !asConfig.ServiceConfig.DisableDaemon {

		log.Println("[Install] Installing Daemon Service...")
		service, ok := environment.Services[config.Daemon.Name]
		if !ok {
			log.Println("[Install] Dispatcher Service not found in environment, name mismatch...")
			return fmt.Errorf("daemon Service not found in environment, name mismatch")
		}

		err = service.Install()
		if err != nil {
			return err
		}
		log.Println("[Install] Installed Daemon Service")
	}

	controlServices := environment.GetControlServices()
	for _, service := range controlServices {
		log.Println("[Install] Installing Control Service: ", service.Name)
		err := service.Install()
		if err != nil {
			return err
		}
		log.Println("[Install] Installed Control Service: ", service.Name)
	}

	borderRouters := environment.GetBorderRouters()
	for _, service := range borderRouters {
		log.Println("[Install] Installing Border Router Service: ", service.Name)
		err := service.Install()
		if err != nil {
			return err
		}
		environment.Services[service.Name] = service
		log.Println("[Install] Installed Border Router Service: ", service.Name)
	}

	log.Println("[Install] Installing scion-orchestrator Service")
	service, ok := environment.Services["scion-orchestrator"]
	if !ok {
		log.Println("[Install] SCION AS Service not found in environment, name mismatch...")
		return fmt.Errorf("orchestrator service found in environment, name mismatch")
	}

	err = service.Install()
	if err != nil {
		return err
	}
	log.Println("[Install] scion-orchestrator Service installed")

	err = environment.StartAllServices()
	if err != nil {
		return err
	}

	log.Println("[Main] All Services started, waiting for health check")
	// TODO: Health check
	time.Sleep(5 * time.Second)
	if !environment.UpdateHealthCheck() {
		log.Println("[Main] Not all services started properly, see the details")
		jsonStatus, _ := metrics.Status.Json()
		fmt.Printf("%s", string(jsonStatus))

		return fmt.Errorf("not all services started properly, Please check the logs or try again")
	} else {
		jsonStatus, _ := metrics.Status.Json()
		fmt.Printf("%s", string(jsonStatus))
	}

	return nil
}
