package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/jessevdk/go-flags"
	"golang.org/x/sync/errgroup"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
	"github.com/netsys-lab/scion-as/pkg/ascrypto"
	"github.com/netsys-lab/scion-as/pkg/bootstrap"
	"github.com/netsys-lab/scion-as/pkg/fileops"
	"github.com/netsys-lab/scion-as/pkg/metrics"
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
		scionConfig, err = conf.LoadSCIONConfig()
		if err != nil {
			log.Println("[Main] Error loading scion config: ", err)
			log.Fatal(err)
		}
		// log.Println("[Main] Config loaded")
		log.Println(scionConfig.Log())
	}

	env := environment.HostEnv

	install := len(args) > 0 && args[0] == "install"
	run := len(args) > 0 && args[0] == "run"

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	if opts.Config != "" { // Run as a service
		log.Println("[Main] Running as service")
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
		}()
		err = runStandalone(env)
	} else if install {
		log.Println("[Main] Installing as service")
		err = runInstall(env, scionConfig)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func runBackgroundServices(env *environment.HostEnvironment, config *conf.Config) error {
	log.Println("[Main] Running background services")

	var eg errgroup.Group

	eg.Go(func() error {
		return bootstrap.RunBootstrapService(env.ConfigPath, config.Bootstrap.Server)
	})

	eg.Go(func() error {
		return metrics.RunStatusHTTPServer(config.Metrics.Server)
	})

	eg.Go(func() error {
		healtchCheck := environment.NewServiceHealthCheck()
		healtchCheck.Run()
		return nil
	})

	eg.Go(func() error {
		// TODO: Obtain ISD AS from config
		renewer := ascrypto.NewCertificateRenewer(env.ConfigPath, config.IsdAs, 6)
		renewer.Run()
		return nil
	})

	return eg.Wait()
}

func runService(env *environment.HostEnvironment, config *conf.Config) error {
	//log.Println("[Main] Running service")
	//time.Sleep(30 * time.Second)
	return nil
}
