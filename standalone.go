package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
	"github.com/netsys-lab/scion-as/pkg/bootstrap"
	"github.com/netsys-lab/scion-as/pkg/fileops"
	"golang.org/x/sync/errgroup"
)

func runStandalone(env *environment.HostEnvironment, config *conf.Config) error {

	// log.Println("[Main] Running standalone")
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

	serviceCount := 0

	if config.Mode == "endhost" {
		log.Println("[Main] Running bootstrapper to fetch configuration...")
		err = bootstrap.BootstrapFromAddress(env, config)
		if err != nil {
			log.Println("[Main] Failed to bootstrap host: ", err)
			log.Fatal(err)
		}
	}

	eg := errgroup.Group{}

	if scionConfig.Dispatcher != nil && !config.ServiceConfig.DisableDispatcher {
		serviceCount++
		eg.Go(func() error {
			log.Println("[Main] Running scion dispatcher")
			err := runStandaloneDispatcher(*env, *scionConfig.Dispatcher)
			if err != nil {
				log.Println("[Main] Error running dispatcher: ", err)
				return err
			}
			return nil
		})
		// TODO: CHeck if dispatcher is running
		time.Sleep(2 * time.Second)
	}

	if !config.ServiceConfig.DisableDaemon {
		serviceCount++
		eg.Go(func() error {
			log.Println("[Main] Running scion daemon")
			err := runStandaloneDaemon(*env, scionConfig.Daemon)
			if err != nil {
				log.Println("[Main] Error running daemon: ", err)
				return err
			}
			return nil
		})
	}

	for _, service := range scionConfig.BorderRouters {
		serviceCount++
		func(service conf.SCIONService) {
			eg.Go(func() error {
				log.Println("[Main] Running router: ", service.Name)
				err := runStandaloneRouter(*env, service)
				if err != nil {
					log.Println("[Main] Error running router: ", err)
					return err
				}
				return nil
			})
		}(service)
	}

	for _, service := range scionConfig.ControlServices {
		serviceCount++
		func(service conf.SCIONService) {
			eg.Go(func() error {
				log.Println("[Main] Running control: ", service.Name)
				err := runStandaloneControlService(*env, service)
				if err != nil {
					log.Println("[Main] Error running control: ", err)
					return err
				}
				return nil
			})
		}(service)
	}

	if serviceCount == 0 {
		log.Println("[Main] No main services to run, keeping app alive to run background services")
		eg.Go(func() error {
			// Wait forever
			select {}
		})
	}

	log.Println("[Main] Waiting for shutdown...")
	err = eg.Wait()

	if err != nil {
		environment.KillAllChilds()
		log.Fatal(err)
	}

	return nil
}

func runStandaloneRouter(env environment.HostEnvironment, service conf.SCIONService) error {

	tmpRouterFile := filepath.Join(env.TmpConfigPath, fmt.Sprintf("br%d-tmp.toml", service.Index))
	err := fileops.CopyFile(tmpRouterFile, filepath.Join(env.ConfigPath, service.ConfigFile))
	if err != nil {
		return err
	}
	err = fileops.ReplaceStringInFile(tmpRouterFile, "{configDir}", env.DaemonConfigPath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder for router config in " + tmpRouterFile + ": " + err.Error())
	}

	err = fileops.ReplaceSingleBackslashWithDouble(tmpRouterFile)
	if err != nil {
		log.Fatal(err)
	}

	router := environment.StandaloneService{
		Name:       service.Name,
		BinaryPath: filepath.Join("bin", "router"),
		ConfigPath: tmpRouterFile,
		Logfile:    filepath.Join("config", "logs", fmt.Sprintf("br-%d.log", service.Index)),
	}

	return router.Run()
}

func runStandaloneControlService(env environment.HostEnvironment, service conf.SCIONService) error {

	tmpControlFile := filepath.Join(env.TmpConfigPath, fmt.Sprintf("cs%d-tmp.toml", service.Index))
	err := fileops.CopyFile(tmpControlFile, filepath.Join(env.ConfigPath, service.ConfigFile))
	if err != nil {
		return err
	}
	err = fileops.ReplaceStringInFile(tmpControlFile, "{configDir}", env.DaemonConfigPath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder for router config in " + tmpControlFile + ": " + err.Error())
	}

	err = fileops.ReplaceStringInFile(tmpControlFile, "{databaseDir}", env.DatabasePath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder for router config in " + tmpControlFile + ": " + err.Error())
	}

	err = fileops.ReplaceSingleBackslashWithDouble(tmpControlFile)
	if err != nil {
		log.Fatal(err)
	}

	control := environment.StandaloneService{
		Name:       service.Name,
		BinaryPath: filepath.Join("bin", "control"),
		ConfigPath: tmpControlFile,
		Logfile:    filepath.Join("config", "logs", fmt.Sprintf("cs-%d.log", service.Index)),
	}

	return control.Run()
}

func runStandaloneDaemon(env environment.HostEnvironment, service conf.SCIONService) error {

	tmpDaemonFile := filepath.Join(env.TmpConfigPath, "sciond-tmp.toml")
	err := fileops.CopyFile(tmpDaemonFile, filepath.Join(env.ConfigPath, service.ConfigFile))
	if err != nil {
		return err
	}
	err = fileops.ReplaceStringInFile(tmpDaemonFile, "{configDir}", env.DaemonConfigPath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder for daemon config in " + tmpDaemonFile + ": " + err.Error())
	}

	err = fileops.ReplaceStringInFile(tmpDaemonFile, "{databaseDir}", env.DatabasePath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder for daemon config in " + tmpDaemonFile + ": " + err.Error())
	}

	err = fileops.ReplaceSingleBackslashWithDouble(tmpDaemonFile)
	if err != nil {
		log.Fatal(err)
	}

	daemon := environment.StandaloneService{
		Name:       service.Name,
		BinaryPath: filepath.Join("bin", "daemon"),
		ConfigPath: tmpDaemonFile,
		Logfile:    filepath.Join("config", "logs", "sciond.log"),
	}

	return daemon.Run()
}

func runStandaloneDispatcher(env environment.HostEnvironment, service conf.SCIONService) error {

	tmpDispatcherFile := filepath.Join(env.TmpConfigPath, "dispatcher-tmp.toml")
	err := fileops.CopyFile(tmpDispatcherFile, filepath.Join(env.ConfigPath, service.ConfigFile))
	if err != nil {
		return err
	}

	if fileops.FileOrFolderExists("/run/shm/dispatcher/default.sock") {
		os.Remove("/run/shm/dispatcher/default.sock")
	}

	dispatcher := environment.StandaloneService{
		Name:       service.Name,
		BinaryPath: filepath.Join("bin", "dispatcher"),
		ConfigPath: tmpDispatcherFile,
		Logfile:    filepath.Join("config", "logs", "dispatcher.log"),
	}

	return dispatcher.Run()

}
