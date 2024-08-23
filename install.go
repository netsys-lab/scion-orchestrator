package main

import (
	"log"
	"os"

	"github.com/netsys-lab/scion-as/environment"
)

func runInstall(env *environment.HostEnvironment) error {

	// TODO: Proper error handling, do not fatal in here...
	// TODO: Mcast Bootstrapping, and all the other things too
	err := os.MkdirAll(env.ConfigPath, 0777)
	if err != nil {
		return err
	}

	log.Println("[Main] Installing files to ", env.BasePath)
	err = env.Install()
	if err != nil {
		return err
	}

	/*binPath := "/usr/bin/"

	// TODO: Windows and MacOS Support!
	switch runtime.GOOS {
	case "linux":
		break
	}

	log.Println("[Main] Installing System Service")
	service := &environment.SystemService{
		Name:       "scion-as",
		BinaryPath: filepath.Join(binPath, "scion-as"),
		ConfigPath: filepath.Join(env.ConfigPath, "sciond.toml"),
	}

	err = service.Install()
	if err != nil {
		return err
	}
	log.Println("[Main] Service installed")

	err = service.Start()
	if err != nil {
		return err
	}
	log.Println("[Main] Service started")

	for {
		if service.IsRunning() {
			break
		}
		time.Sleep(3 * time.Second)
	}
	log.Println("[Main] Service is running, closing this one now...")
	*/
	return nil
}
