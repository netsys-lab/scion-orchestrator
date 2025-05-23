package apiv1

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/environment"
	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
	"github.com/netsys-lab/scion-orchestrator/pkg/netutils"
	"github.com/netsys-lab/scion-orchestrator/pkg/scionutils"
)

func DoInstallHandler(eng *gin.RouterGroup, env *environment.HostEnvironment, scionConfig *conf.SCIONConfig, asConfig *conf.Config) {

	eng.POST("install", func(c *gin.Context) {

		var setup environment.InstallSetup
		err := c.BindJSON(&setup)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON"})
			return
		}

		err = runInstall(env, asConfig, scionConfig, &setup)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Installation successful"})
		log.Println("[Install] Installation successful, shutting down")
		go func() {
			time.Sleep(2000 * time.Millisecond)
			os.Exit(0)
		}()
	})
}

func runInstall(env *environment.HostEnvironment, asConfig *conf.Config, scionConfig *conf.SCIONConfig, setup *environment.InstallSetup) error {

	// TODO: Binary copy does not work when services are running

	// TODO: Proper error handling, do not fatal in here...
	// TODO: Mcast Bootstrapping, and all the other things too

	if setup.InstallDir == "" {
		return fmt.Errorf("InstallDir is empty")
	}
	if setup.AdminUsername == "" || setup.AdminPassword == "" {
		return fmt.Errorf("AdminUsername or AdminPassword is empty")
	}

	if setup.DeployBorderRouter {
		if setup.BorderRouterAddr == "" {
			return fmt.Errorf("BorderRouterAddr is empty despite having DeployBorderRouter set to true")
		}

		udpAddr, err := net.ResolveUDPAddr("udp", setup.BorderRouterAddr)
		if err != nil {
			return fmt.Errorf("Invalid local IP %s", setup.BorderRouterAddr)

		}

		isValidLink, err := netutils.IsLocalIPWithMTU(udpAddr.IP.String(), 1400) // Default MTU, should always work
		if !isValidLink && err == nil {
			return fmt.Errorf("Local IP %s not found on this host or MTU exceeds interface MTU", setup.BorderRouterAddr)
		}
		if err != nil {
			fmt.Println("IP %s is invalid on this host", setup.BorderRouterAddr, "error: ", err)
			return fmt.Errorf("Could not detect if IP %s is valid on this host", setup.BorderRouterAddr)
		}

		if !netutils.IsUDPPortFree(setup.BorderRouterAddr) {
			return fmt.Errorf("UDP port %s is already in use", setup.BorderRouterAddr)
		}
	}

	if setup.DeployControl {
		if setup.ControlAddr == "" {
			return fmt.Errorf("ControlAddr is empty despite having DeployControl set to true")
		}

		tcpAddr, err := net.ResolveTCPAddr("tcp", setup.ControlAddr)
		if err != nil {
			return fmt.Errorf("Invalid local IP %s", setup.ControlAddr)

		}

		isValidLink, err := netutils.IsLocalIPWithMTU(tcpAddr.IP.String(), 1400) // Default MTU, should always work
		if !isValidLink && err == nil {
			return fmt.Errorf("Local IP %s not found on this host or MTU exceeds interface MTU", setup.ControlAddr)
		}
		if err != nil {
			fmt.Println("IP %s is invalid on this host", setup.ControlAddr, "error: ", err)
			return fmt.Errorf("Could not detect if IP %s is valid on this host", setup.ControlAddr)
		}

		if !netutils.IsTCPPortFree(setup.ControlAddr) {
			return fmt.Errorf("UDP port %s is already in use", setup.ControlAddr)
		}
	}

	if !scionutils.IsValidISDAS(setup.ISDAs) {
		return fmt.Errorf("Invalid ISD-AS %s", setup.ISDAs)
	}

	if setup.InstallDir != "" {
		env.SetConfigPath(setup.InstallDir)
	}

	err := os.MkdirAll(env.ConfigPath, 0777)
	if err != nil {
		return err
	}

	err = environment.LoadServices(env, scionConfig, asConfig)
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
	err = env.Install(setup)
	if err != nil {
		return err
	}

	// Update config with username password
	newConfig, err := conf.LoadConfig(filepath.Join(env.ConfigPath, "scion-orchestrator.toml"))
	if err != nil {
		return err
	}

	newConfig.Api.Users = make([]string, 0)
	newConfig.Api.Users = append(newConfig.Api.Users, fmt.Sprintf("%s:%s", setup.AdminUsername, setup.AdminPassword))

	err = newConfig.Save(filepath.Join(env.ConfigPath, "scion-orchestrator.toml"))
	if err != nil {
		return err
	}

	if scionConfig.Dispatcher != nil && !asConfig.ServiceConfig.DisableDispatcher {
		log.Println("[Install] Installing Dispatcher Service...")
		service, ok := environment.Services[scionConfig.Dispatcher.Name]
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
		service, ok := environment.Services[scionConfig.Daemon.Name]
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
