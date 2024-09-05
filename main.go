package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
	"golang.org/x/sync/errgroup"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
	"github.com/netsys-lab/scion-as/pkg/bootstrap"
	"github.com/netsys-lab/scion-as/pkg/fileops"
	"github.com/netsys-lab/scion-as/pkg/metrics"
	"github.com/netsys-lab/scion-as/pkg/scionca"
)

var opts struct {
	// Example of verbosity with level
	// Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`

	// Example of optional value
	Config string `short:"c" long:"config" description:"Path to configuration file"`

	// Example of optional value
	InstallDir string `short:"d" long:"dir" description:"Directory to install scion-as as service" `

	// Example of optional value
	// Install bool `short:"i" long:"install" description:"Install scion-as as a system service" `
}

var scionConfig *conf.SCIONConfig

var install bool
var run bool
var shutdown bool

func main() {

	args, err := flags.Parse(&opts)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("[Main] Starting scion-as")
	log.Println("[Main] Running on ", runtime.GOOS)
	log.Println("[Main] Args: ", args)

	configPath := filepath.Join("config", "scion-as.toml")
	if opts.Config != "" {
		configPath = opts.Config
	}

	config, err := conf.LoadConfig(configPath)
	log.Println("[Main] scion-as config loaded successfully")

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

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	if opts.Config != "" { // Run as a service
		log.Println("[Main] Running as service")

		if config.Mode == "endhost" {
			log.Println("[Main] Running bootstrapper to fetch configuration...")
			err = bootstrap.BootstrapFromAddress(env, config)
			if err != nil {
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
		go func() {
			err = runService(env, config)
			if err != nil {
				log.Fatal(err)
			}
		}()

		log.Println("[Main] SCION-AS Service started, waiting for termination signal")
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
		jsonStatus, _ := metrics.ASStatus.Json()
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

func runBackgroundServices(env *environment.HostEnvironment, config *conf.Config) error {
	log.Println("[Main] Running background services")
	var eg errgroup.Group

	if config.Mode == "endhost" {
		eg.Go(func() error {
			return metrics.RunStatusHTTPServer(config.Metrics.Server)
		})

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
				// TODO: Obtain ISD AS from config
				renewer := scionca.NewCertificateRenewer(env.ConfigPath, config.IsdAs, 6)
				renewer.Run()
				return nil
			})
		}
	} else {
		if !config.ServiceConfig.DisableBootstrapServer {
			eg.Go(func() error {
				return bootstrap.RunBootstrapService(env.ConfigPath, config.Bootstrap.Server)
			})
		}

		eg.Go(func() error {
			return metrics.RunStatusHTTPServer(config.Metrics.Server)
		})

		log.Println("[Main] Running background services for CA")
		// log.Println(config.Ca.Clients)
		if config.Ca.Clients != nil && len(config.Ca.Clients) > 0 {
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
				renewer.Run()
				return nil
			})
		}
	}

	return eg.Wait()
}

func runService(env *environment.HostEnvironment, config *conf.Config) error {
	//log.Println("[Main] Running service")
	//time.Sleep(30 * time.Second)

	return nil
}
