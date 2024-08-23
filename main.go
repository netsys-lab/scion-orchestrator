package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jessevdk/go-flags"
	"golang.org/x/sync/errgroup"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
	"github.com/netsys-lab/scion-as/pkg/bootstrap"
	"github.com/netsys-lab/scion-as/pkg/fileops"
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

	if opts.Config != "" { // Run as a service
		log.Println("[Main] Running as service")
		go func() {
			err = runBackgroundServices(env, config)
			if err != nil {
				log.Fatal(err)
			}
		}()

		err = runService(env, config)
	} else if run {
		log.Println("[Main] Running in standalone mode")
		go func() {
			err = runBackgroundServices(env, config)
			if err != nil {
				log.Fatal(err)
			}
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

	eg := errgroup.Group{}

	eg.Go(func() error {
		return bootstrap.RunTrcFileWatcher(env.ConfigPath)
	})

	eg.Go(func() error {
		return bootstrap.RunBootstrapServer(env.ConfigPath, config.Bootstrap.Server)
	})

	err := eg.Wait()

	return err
}

func runService(env *environment.HostEnvironment, config *conf.Config) error {

	return nil
}
