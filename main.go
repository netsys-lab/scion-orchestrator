package main

import (
	"context"
	"errors"
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
		err = runService(endhostEnv, config)
	} else if run {
		log.Println("Running in standalone mode")
		err = runStandalone(endhostEnv)
	} else if install {
		log.Println("Installing as service")
		err = runInstall(endhostEnv)
	}

	if err != nil {
		log.Fatal(err)
	}
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

func runStandalone(env *environment.EndhostEnvironment) error {
	var wg sync.WaitGroup

	// log.Println("Running standalone")
	env.ChangeToStandalone()

	err := os.MkdirAll(env.DatabasePath, 0777)
	if err != nil {
		return err
	}

	err = os.MkdirAll(env.LogPath, 0777)
	if err != nil {
		return err
	}

	err = os.MkdirAll(env.TmpConfigPath, 0777)
	if err != nil {
		return err
	}

	tmpSciondFile := filepath.Join(env.TmpConfigPath, "sciond-tmp.toml")

	err = fileops.CopyFile(tmpSciondFile, filepath.Join(env.ConfigPath, "sciond.toml"))
	if err != nil {
		return err
	}
	err = fileops.ReplaceStringInFile(tmpSciondFile, "{configDir}", env.DaemonConfigPath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder in sciond.toml: " + err.Error())
	}

	err = fileops.ReplaceStringInFile(tmpSciondFile, "{databaseDir}", env.DatabasePath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder in sciond.toml: " + err.Error())
	}

	wg.Add(1)
	go func() {

		// sciondFile := filepath.Join("config", "sciond.toml")
		defer wg.Done()
		log.Println("Running scion deamon with config file: ", tmpSciondFile)
		err := runDaemon(context.Background(), tmpSciondFile)
		if err != nil {
			log.Println("Error running daemon: ", err)
			environment.KillAllChilds()
			log.Fatal(err)
		}
		log.Println("Daemon running")
	}()

	for _, service := range scionConfig.BorderRouters {
		wg.Add(1)
		log.Println("Running router: ", service.Name)
		go func(service conf.SCIONService) {
			defer wg.Done()
			err := runStandaloneRouter(*env, service)
			if err != nil {
				log.Println("Error running router: ", err)
				environment.KillAllChilds()
				log.Fatal(err)
			}
		}(service)
	}

	for _, service := range scionConfig.ControlServices {
		wg.Add(1)
		log.Println("Running control: ", service.Name)
		go func(service conf.SCIONService) {
			defer wg.Done()
			time.Sleep(3 * time.Second) // TODO: Check uptime
			err := runStandaloneControlService(*env, service)
			if err != nil {
				log.Println("Error running control: ", err)
				environment.KillAllChilds()
				log.Fatal(err)
			}
		}(service)
	}

	/*wg.Add(1)
	go func() {
		// routerFile := filepath.Join("config", "br-1.toml")
		defer wg.Done()
		log.Println("Running scion router with config file: ", tmpRouterFile)
		err := runRouter(context.Background(), tmpRouterFile)
		if err != nil {
			log.Println("Error running router: ", err)
			environment.KillAllChilds()
			log.Fatal(err)
		}
		log.Println("Router running")
	}()*/

	/*tmpControlFile := filepath.Join(env.TmpConfigPath, "cs1-tmp.toml")
	err = fileops.CopyFile(tmpControlFile, filepath.Join(env.ConfigPath, "cs-1.toml"))
	if err != nil {
		return err
	}
	err = fileops.ReplaceStringInFile(tmpControlFile, "{configDir}", env.DaemonConfigPath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder in sciond.toml: " + err.Error())
	}

	err = fileops.ReplaceStringInFile(tmpControlFile, "{databaseDir}", env.DatabasePath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder in sciond.toml: " + err.Error())
	}

	wg.Add(1)
	go func() {
		// controlFile := filepath.Join("config", "cs-1.toml")
		defer wg.Done()
		time.Sleep(2 * time.Second)
		log.Println("Running scion control with config file: ", tmpControlFile)
		err := runControl(context.Background(), tmpControlFile)
		if err != nil {
			log.Println("Error running control: ", err)
			environment.KillAllChilds()
			log.Fatal(err)
		}
		log.Println("Control running")
	}()*/

	tmpDispatcherFile := filepath.Join(env.TmpConfigPath, "dispatcher-tmp.toml")
	err = fileops.CopyFile(tmpDispatcherFile, filepath.Join(env.ConfigPath, "dispatcher.toml"))
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		// controlFile := filepath.Join("config", "cs-1.toml")
		defer wg.Done()
		log.Println("Running scion dispatcher with config file: ", tmpDispatcherFile)
		err := runDispatcher(context.Background(), tmpDispatcherFile)
		if err != nil {
			log.Println("Error running dispatcher: ", err)
			environment.KillAllChilds()
			log.Fatal(err)
		}
		log.Println("Dispatcher running")
	}()

	wg.Wait()
	log.Println("All services running")
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
