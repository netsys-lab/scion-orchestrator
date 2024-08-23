package environment

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/netsys-lab/scion-as/pkg/fileops"
)

var HostEnv *HostEnvironment

type HostEnvironment struct {
	BasePath          string
	ConfigPath        string
	DaemonConfigPath  string
	ControlConfigPath string
	RouterConfigPath  string
	DatabasePath      string
	TmpConfigPath     string
	LogPath           string
}

func (endhostEnv *HostEnvironment) ChangeToStandalone() {
	dir, _ := os.Getwd()
	dir = filepath.Join(dir, "config")
	endhostEnv.BasePath = dir
	endhostEnv.ConfigPath = dir
	endhostEnv.DaemonConfigPath = dir
	endhostEnv.ControlConfigPath = dir
	endhostEnv.RouterConfigPath = dir
	endhostEnv.DatabasePath = filepath.Join(dir, "database")
	endhostEnv.TmpConfigPath = filepath.Join(dir, "tmp")
	endhostEnv.LogPath = filepath.Join(dir, "logs")
}

func (endhostEnv *HostEnvironment) installBinaries() error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		fmt.Println("Copying binaries to ", "/usr/bin")
		err := fileops.CopyFile(filepath.Join("/usr/bin", "scion-as"), filepath.Join(workDir, "scion-as"))
		if err != nil {
			return err
		}

		// Make files executable
		err = os.Chmod(filepath.Join("/usr/bin", "scion-as"), 0777)
		if err != nil {
			return err
		}

		binaries := []string{"scion", "control", "router", "dispatcher", "gateway"}
		for _, binary := range binaries {
			err = fileops.CopyFile(filepath.Join("/usr/bin", binary), filepath.Join(workDir, "bin", binary))
			if err != nil {
				return err
			}
			err = os.Chmod(filepath.Join("/usr/bin", binary), 0777)
			if err != nil {
				return err
			}
		}

	case "darwin":
		return nil
		// return installMacService(serviceName, binaryPath, configPath)
	case "windows":
		return nil
		// return installWindowsService(serviceName, binaryPath, configPath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return nil
}

func (endhostEnv *HostEnvironment) Install() error {
	err := os.MkdirAll(endhostEnv.BasePath, 0777)
	if err != nil {
		return err
	}

	err = HostEnv.installBinaries()
	if err != nil {
		return err
	}

	fmt.Println("Copying config files to ", endhostEnv.ConfigPath)

	destSciondFile := filepath.Join(endhostEnv.ConfigPath, "sciond.toml")
	err = fileops.CopyFile(destSciondFile, filepath.Join("config", "sciond.toml"))
	if err != nil {
		return err
	}

	destBootstrapperFile := filepath.Join(endhostEnv.ConfigPath, "bootstrapper.toml")
	err = fileops.CopyFile(destBootstrapperFile, filepath.Join("config", "bootstrapper.toml"))
	if err != nil {
		return err
	}

	destScionHostFile := filepath.Join(endhostEnv.ConfigPath, "scion-as.toml")
	err = fileops.CopyFile(destScionHostFile, filepath.Join("config", "scion-as.toml"))
	if err != nil {
		return err
	}

	err = fileops.ReplaceStringInFile(destSciondFile, "{configDir}", endhostEnv.DaemonConfigPath)
	if err != nil {
		return errors.New("Failed to configure folder in sciond.toml: " + err.Error())
	}
	err = fileops.ReplaceStringInFile(destBootstrapperFile, "{configDir}", endhostEnv.DaemonConfigPath)
	if err != nil {
		return errors.New("Failed to configure folder in sciond.toml: " + err.Error())
	}

	/*err = fileops.ReplaceStringInFile(filepath.Join(endhostEnv.ConfigPath, "dispatcher.toml"), "{configDir}", endhostEnv.ConfigPath)
	if err != nil {
		log.Fatal("Failed to configure folder in sciond.toml: ", err)
	}*/

	// TODO: This could also be only windows specific
	err = fileops.ReplaceSingleBackslashWithDouble(filepath.Join(endhostEnv.ConfigPath, "sciond.toml"))
	if err != nil {
		log.Fatal(err)
	}

	err = fileops.ReplaceSingleBackslashWithDouble(filepath.Join(endhostEnv.ConfigPath, "bootstrapper.toml"))
	if err != nil {
		log.Fatal(err)
	}
	return nil

}
