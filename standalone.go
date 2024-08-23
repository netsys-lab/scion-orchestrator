package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
	"github.com/netsys-lab/scion-as/pkg/fileops"
)

func runStandalone(env *environment.HostEnvironment) error {
	var wg sync.WaitGroup

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

	if scionConfig.Dispatcher != nil {
		wg.Add(1)
		go func() {
			// sciondFile := filepath.Join("config", "sciond.toml")
			defer wg.Done()
			log.Println("[Main] Running scion dispatcher")
			err := runStandaloneDispatcher(*env, *scionConfig.Dispatcher)
			if err != nil {
				log.Println("[Main] Error running dispatcher: ", err)
				environment.KillAllChilds()
				log.Fatal(err)
			}
			log.Println("[Main] Dispatcher running")
		}()
		// TODO: CHeck if dispatcher is running
		time.Sleep(2 * time.Second)
	}

	wg.Add(1)
	go func() {
		// sciondFile := filepath.Join("config", "sciond.toml")
		defer wg.Done()
		log.Println("[Main] Running scion daemon")
		err := runStandaloneDaemon(*env, scionConfig.Daemon)
		if err != nil {
			log.Println("[Main] Error running daemon: ", err)
			environment.KillAllChilds()
			log.Fatal(err)
		}
		log.Println("[Main] Daemon running")
	}()

	for _, service := range scionConfig.BorderRouters {
		wg.Add(1)
		log.Println("[Main] Running router: ", service.Name)
		go func(service conf.SCIONService) {
			defer wg.Done()
			err := runStandaloneRouter(*env, service)
			if err != nil {
				log.Println("[Main] Error running router: ", err)
				environment.KillAllChilds()
				log.Fatal(err)
			}
		}(service)
	}

	for _, service := range scionConfig.ControlServices {
		wg.Add(1)
		log.Println("[Main] Running control: ", service.Name)
		go func(service conf.SCIONService) {
			defer wg.Done()
			err := runStandaloneControlService(*env, service)
			if err != nil {
				log.Println("[Main] Error running control: ", err)
				environment.KillAllChilds()
				log.Fatal(err)
			}
		}(service)
	}

	wg.Wait()
	log.Println("[Main] All services running")
	return nil
}

func runStandaloneRouter(env environment.HostEnvironment, service conf.SCIONService) error {

	tmpRouterFile := filepath.Join(env.TmpConfigPath, fmt.Sprintf("br%d-tmp.toml", service.Index))
	err := fileops.CopyFile(tmpRouterFile, service.ConfigFile)
	if err != nil {
		return err
	}
	err = fileops.ReplaceStringInFile(tmpRouterFile, "{configDir}", env.DaemonConfigPath+string(os.PathSeparator))
	if err != nil {
		return errors.New("Failed to configure folder for router config in " + tmpRouterFile + ": " + err.Error())
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
	err := fileops.CopyFile(tmpControlFile, service.ConfigFile)
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
	err := fileops.CopyFile(tmpDaemonFile, service.ConfigFile)
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
	err := fileops.CopyFile(tmpDispatcherFile, service.ConfigFile)
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
