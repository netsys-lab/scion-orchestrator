package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/netsys-lab/scion-orchestrator/conf"
	"github.com/netsys-lab/scion-orchestrator/environment"
	"github.com/netsys-lab/scion-orchestrator/pkg/apiv1"
	"github.com/netsys-lab/scion-orchestrator/pkg/bootstrap"
	"github.com/netsys-lab/scion-orchestrator/pkg/certutils"
	"github.com/netsys-lab/scion-orchestrator/pkg/fileops"
	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
	"github.com/netsys-lab/scion-orchestrator/pkg/scionca"
	"github.com/netsys-lab/scion-orchestrator/ui"
	scionpila "github.com/netsys-lab/scion-pila"
)

var opts struct {
	// Example of verbosity with level
	// Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`

	// Example of optional value
	Config string `short:"c" long:"config" description:"Path to configuration file"`

	// Example of optional value
	InstallDir string `short:"d" long:"dir" description:"Directory to install scion-orchestrator as service" `

	// Example of optional value
	// Install bool `short:"i" long:"install" description:"Install scion-orchestrator as a system service" `
}

var scionConfig *conf.SCIONConfig

var install bool
var run bool
var shutdown bool
var restart bool
var installWizard bool

func main() {

	// Check that its just running as standalone without args -> installWizard
	if len(os.Args) == 1 {
		log.Println("[Main] Running as standalone without args")
		log.Println("[Main] Running install wizard")
		installWizard = true
	}
	args, err := flags.Parse(&opts)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("[Main] Starting scion-orchestrator")
	log.Println("[Main] Running on ", runtime.GOOS)
	log.Println("[Main] Args: ", args)

	configPath := filepath.Join("config", "scion-orchestrator.toml")
	if opts.Config != "" {
		configPath = opts.Config
	}

	config, err := conf.LoadConfig(configPath)
	if err != nil {
		log.Fatal("[Main] failed to load config ", config)
	}
	log.Println("[Main] scion-orchestrator config loaded successfully")

	if fileops.FileOrFolderExists("config") {
		// log.Println("[Main] Config folder exists")
		scionConfig, err = conf.LoadSCIONConfig("")
		if err != nil {
			log.Println("[Main] Error loading scion config: ", err)
			log.Fatal(err)
		}
		// log.Println("[Main] Config loaded")
		log.Println(scionConfig.Log())
	} else {
		// Get base path
		scionConfigDir := filepath.Dir(opts.Config)
		scionConfig, err = conf.LoadSCIONConfig(scionConfigDir)
		if err != nil {
			log.Println("[Main] Error loading scion config: ", err)
			log.Fatal(err)
		}
		// log.Println("[Main] Config loaded")
		log.Println(scionConfig.Log())
	}

	env := environment.HostEnv

	install = len(args) > 0 && args[0] == "install"
	run = len(args) > 0 && args[0] == "run"
	shutdown = len(args) > 0 && args[0] == "shutdown"
	restart = len(args) > 0 && args[0] == "restart"

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	metrics.Init()

	if run {
		metrics.Status.ServiceMode = "standalone"
	} else {
		metrics.Status.ServiceMode = "service"
	}

	if installWizard {
		env.ChangeToStandalone()
		log.Println("[Main] Starting API server...")

		go func() {
			sig := <-cancelChan
			log.Printf("[Signal] Caught signal %v", sig)
			environment.KillAllChilds()
			log.Fatal("[Main] Shutting down...")
		}()

		err := runUIApi(env, config)
		if err != nil {
			log.Fatal(err)
		}
	}
	if opts.Config != "" { // Run as a service
		log.Println("[Main] Running as service")

		if config.Mode == "endhost" {
			log.Println("[Main] Running bootstrapper to fetch configuration...")
			err = bootstrap.BootstrapFromAddress(env, config)
			if err != nil && !config.Bootstrap.AllowClientFail {
				log.Println("[Main] Failed to bootstrap host: ", err)
				log.Fatal(err)
			}
		}

		err := environment.LoadServices(env, scionConfig, config)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			err = runBackgroundServices(env, config)
			if err != nil {
				log.Fatal(err)
			}
		}()

		log.Println("[Main] scion-orchestrator Service started, waiting for termination signal")
		sig := <-cancelChan
		log.Printf("[Signal] Caught signal %v", sig)

		// TODO: Add proper service initialization here
		// time.Sleep(10 * time.Minute)
	} else if run {
		log.Println("[Main] Running in standalone mode")
		go func() {
			err = runBackgroundServices(env, config)
			if err != nil {
				log.Fatal(err)
			}
		}()
		go func() {
			sig := <-cancelChan
			log.Printf("[Signal] Caught signal %v", sig)
			environment.KillAllChilds()
			log.Fatal("[Main] Shutting down...")
		}()
		err = runStandalone(env, config)
	} else if install {
		log.Println("[Main] Installing as service")
		err = runInstall(env, scionConfig, config)
	} else if shutdown {
		// log.Println("[Main] Shutting down all service")
		err = runShutdown(env, scionConfig, *config)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("[Main] All services stopped. AS Status is now:")
		jsonStatus, _ := metrics.Status.Json()
		fmt.Printf("%s", string(jsonStatus))
	} else if restart {
		// log.Println("[Main] Shutting down all service")
		err = runRestart(env, scionConfig, *config)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("[Main] All services restarted. AS Status is now:")
		jsonStatus, _ := metrics.Status.Json()
		fmt.Printf("%s", string(jsonStatus))
	}

	if err != nil {
		log.Fatal(err)
	}
}

func runShutdown(env *environment.HostEnvironment, config *conf.SCIONConfig, asConfig conf.Config) error {
	log.Println("[Main] Shutting down SCION AS")
	err := environment.LoadServices(env, config, &asConfig)
	if err != nil {
		return err
	}

	log.Println("[Main] Stopping all services")
	err = environment.StopAllServices()
	if err != nil {
		return err
	}
	return nil
}

func runRestart(env *environment.HostEnvironment, config *conf.SCIONConfig, asConfig conf.Config) error {
	log.Println("[Main] Restarting SCION AS")
	err := environment.LoadServices(env, config, &asConfig)
	if err != nil {
		return err
	}

	log.Println("[Main] Stopping all services")
	err = environment.StopAllServices()
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	log.Println("[Main] Starting all services")
	err = environment.StartAllServices()
	if err != nil {
		return err
	}
	return nil
}

func runBackgroundServices(env *environment.HostEnvironment, config *conf.Config) error {
	log.Println("[Main] Running background services")
	var eg errgroup.Group

	eg.Go(func() error {
		return metrics.RunStatusHTTPServer(config.Metrics.Server)
	})

	eg.Go(func() error {
		return metrics.RunPrometheusHTTPServer(config.Metrics.Prometheus)
	})

	if len(config.Api.Users) == 0 {
		log.Println("[Main] Warning: No users defined for HTTP access, API and UI can not be accessed")
	} else {
		log.Println("[Main] Starting API server...")
		eg.Go(func() error {
			return runUIApi(env, config)
		})
	}

	if config.Mode == "endhost" {

		// Standalone does its own health check
		if !run {
			eg.Go(func() error {
				healtchCheck := environment.NewServiceHealthCheck()
				healtchCheck.Run()
				return nil
			})
		}
	} else {
		if !config.ServiceConfig.DisableBootstrapServer {
			eg.Go(func() error {
				return bootstrap.RunBootstrapService(env.ConfigPath, config.Bootstrap.Server, config)
			})
		}

		if !config.ServiceConfig.DisablePilaServer && config.Pila.Server != "" {
			fmt.Println(config.Pila)
			eg.Go(func() error {
				log.Println("[Main] Starting PILA server for endhost certificates...")
				scionPilaConfig := &scionpila.SCIONPilaConfig{
					Server:         config.Pila.Server,
					CAKeyPath:      config.Pila.CAKeyPath,
					CACertPath:     config.Pila.CACertPath,
					AllowedSubnets: config.Pila.AllowedSubnets,
				}
				server := scionpila.NewSCIONPilaServer(scionPilaConfig)
				return server.Run()
			})
		}

		log.Println("[Main] Running background services for CA")

		// TODO: Check which services do really need a control plane cert/key
		// TODO: Need to fail here??
		if len(scionConfig.ControlServices) > 0 {
			err := certutils.EnsureASPrivateKeyExists(env.ConfigPath, config.IsdAs)
			if err != nil {
				log.Println("[Main] Error ensuring AS private key exists: ", err)
			}
		}

		// log.Println(config.Ca.Clients)
		if len(config.Ca.Clients) > 0 {
			eg.Go(func() error {
				// TODO: Only run if core AS
				parts := strings.Split(config.IsdAs, "-")
				ca := scionca.NewSCIONCertificateAuthority(env.ConfigPath, parts[0], config.Ca.CertValidityHours)
				err := ca.LoadCA()
				if err != nil {
					return err
				}

				apiServer := scionca.NewCaApiServer(env.ConfigPath, &config.Ca, ca)
				err = apiServer.LoadClientsAndSecrets()
				if err != nil {
					return err
				}

				return apiServer.Run()
			})
		}

		// Standalone does its own health check
		if !run {
			eg.Go(func() error {
				healtchCheck := environment.NewServiceHealthCheck()
				healtchCheck.Run()
				return nil
			})
		}

		if !config.ServiceConfig.DisableCertRenewal {
			eg.Go(func() error {
				renewer := scionca.NewCertificateRenewer(env.ConfigPath, config.IsdAs, 6)
				log.Println("[Main] Starting certificate renewal with 30s delay...")
				time.Sleep(30 * time.Second)
				renewer.Run()
				return nil
			})
		}
	}

	return eg.Wait()
}

func runUIApi(env *environment.HostEnvironment, config *conf.Config) error {
	r := gin.Default()
	// Start Gin server with the newly generated leaf certificate
	// TODO: locate this file properly
	f, _ := os.Create(filepath.Join(env.LogPath, "gin.log"))
	gin.DefaultWriter = io.MultiWriter(f)

	// })

	// eg.Go(func() error {
	err := apiv1.RegisterRoutes(env, config, r, installWizard)
	if err != nil {
		log.Println("[Main] Error registering API routes: ", err)
		return err
	}
	// })

	// eg.Go(func() error {
	err = ui.RegisterRoutes(env, config, r, installWizard)
	if err != nil {
		log.Println("[Main] Error registering UI routes: ", err)
		return err
	}

	apiAddress := ":8443"
	if config.Api.Address != "" {
		apiAddress = config.Api.Address
	}

	log.Println("[Main] Starting API und UI server with TLS on", apiAddress)
	certFile, keyFile, err := apiv1.SetupCertificates(env)
	if err != nil {
		log.Println("[Main] Error creating TLS Certificates of API and UI: ", err)
		return err
	}

	// Start the server with the new certificate and private key
	err = r.RunTLS(apiAddress, certFile, keyFile)
	if err != nil {
		log.Println("[Main] Failed to start TLS API/UI server: ", err)
		return err
	}

	return nil
}
