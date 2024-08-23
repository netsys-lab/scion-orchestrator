package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
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

	log.Println("Starting scion-as")
	log.Println("Running on ", runtime.GOOS)
	log.Println("Args: ", args)

	if fileops.FileOrFolderExists("config") {
		log.Println("Config folder exists")
		scionConfig, err = conf.LoadSCIONConfig()
		if err != nil {
			log.Println("Error loading scion config: ", err)
			log.Fatal(err)
		}
		log.Println("Config loaded")
		log.Println(scionConfig.Log())
	}

	endhostEnv := environment.EndhostEnv

	install := len(args) > 0 && args[0] == "install"
	run := len(args) > 0 && args[0] == "run"

	if opts.Config != "" { // Run as a service
		log.Println("Running as service")
		config, err := conf.LoadConfig(opts.Config)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			err = runBackgroundServices()
			if err != nil {
				log.Fatal(err)
			}
		}()

		err = runService(endhostEnv, config)
	} else if run {
		log.Println("Running in standalone mode")
		go func() {
			err = runBackgroundServices()
			if err != nil {
				log.Fatal(err)
			}
		}()

		err = runStandalone(endhostEnv)
	} else if install {
		log.Println("Installing as service")
		err = runInstall(endhostEnv)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func runBackgroundServices() error {
	log.Println("Running background services")
	return nil
}

func runService(env *environment.EndhostEnvironment, config *conf.Config) error {
	var wg sync.WaitGroup

	// log.Println("Running as service")

	sciondFile := filepath.Join(env.ConfigPath, "sciond.toml")

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Running scion deamon with config file: ", sciondFile)
		err := runDaemon(context.Background(), sciondFile)
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
	return nil
}

func runInstall(env *environment.EndhostEnvironment) error {

	// TODO: Proper error handling, do not fatal in here...
	// TODO: Mcast Bootstrapping, and all the other things too
	err := os.MkdirAll(env.ConfigPath, 0777)
	if err != nil {
		return err
	}

	log.Println("Installing files to ", env.BasePath)
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

	log.Println("Installing System Service")
	service := &environment.SystemService{
		Name:       "scion-as",
		BinaryPath: filepath.Join(binPath, "scion-as"),
		ConfigPath: filepath.Join(env.ConfigPath, "sciond.toml"),
	}

	err = service.Install()
	if err != nil {
		return err
	}
	log.Println("Service installed")

	err = service.Start()
	if err != nil {
		return err
	}
	log.Println("Service started")

	for {
		if service.IsRunning() {
			break
		}
		time.Sleep(3 * time.Second)
	}
	log.Println("Service is running, closing this one now...")

	return nil
}

func runControl(ctx context.Context, configFile string) error {

	controlService := environment.StandaloneService{
		Name:       "control",
		BinaryPath: filepath.Join("bin", "control"),
		ConfigPath: configFile,
		Logfile:    filepath.Join("config", "logs", "control.log"),
	}

	return controlService.Run()
}

func runDaemon(ctx context.Context, configFile string) error {

	controlService := environment.StandaloneService{
		Name:       "daemon",
		BinaryPath: filepath.Join("bin", "daemon"),
		ConfigPath: configFile,
		Logfile:    filepath.Join("config", "logs", "daemon.log"),
	}

	return controlService.Run()
}

func runDispatcher(ctx context.Context, configFile string) error {

	controlService := environment.StandaloneService{
		Name:       "dispatcher",
		BinaryPath: filepath.Join("bin", "dispatcher"),
		ConfigPath: configFile,
		Logfile:    filepath.Join("config", "logs", "dispatcher.log"),
	}

	return controlService.Run()
}

func runRouter(ctx context.Context, configFile string) error {

	controlService := environment.StandaloneService{
		Name:       "router",
		BinaryPath: filepath.Join("bin", "router"),
		ConfigPath: configFile,
		Logfile:    filepath.Join("config", "logs", "router.log"),
	}

	return controlService.Run()
}
