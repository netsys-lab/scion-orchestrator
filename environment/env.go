package environment

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/netsys-lab/scion-orchestrator/pkg/fileops"
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

	// Check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Println("[ChangeToStandalone] Directory does not exist, creating ", dir)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

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

	binPath := "/usr/bin/"

	// TODO: Windows and MacOS Support!
	switch runtime.GOOS {
	case "linux":
		break
	case "darwin":
		binPath = "/usr/local/bin"
		break
	}

	switch runtime.GOOS {
	case "linux":
		log.Println("[Install] Copying binaries to ", binPath)
		err := fileops.CopyFile(filepath.Join(binPath, "scion-orchestrator"), filepath.Join(workDir, "scion-orchestrator"))
		if err != nil {
			return err
		}

		// Make files executable
		err = os.Chmod(filepath.Join(binPath, "scion-orchestrator"), 0777)
		if err != nil {
			return err
		}

		files, err := fileops.ListFilesByPrefixAndSuffix(filepath.Join(workDir, "bin"), "", "")
		if err != nil {
			return err
		}
		binaries := []string{}
		for _, file := range files {
			binaries = append(binaries, filepath.Base(file))
		}
		// binaries := []string{"scion", "control", "router", "dispatcher", "gateway", "daemon"}
		for _, binary := range binaries {
			log.Println("[Install] Copy binary ", filepath.Join(workDir, "bin", binary), "to ", filepath.Join(binPath, binary))
			err = fileops.CopyFile(filepath.Join(binPath, binary), filepath.Join(workDir, "bin", binary))
			if err != nil {
				return err
			}
			err = os.Chmod(filepath.Join(binPath, binary), 0777)
			if err != nil {
				return err
			}
		}

	case "darwin":
		log.Println("[Install] Copying binaries to ", binPath)
		err := fileops.CopyFile(filepath.Join(binPath, "scion-orchestrator"), filepath.Join(workDir, "scion-orchestrator"))
		if err != nil {
			return err
		}

		// Make files executable
		err = os.Chmod(filepath.Join(binPath, "scion-orchestrator"), 0777)
		if err != nil {
			return err
		}

		files, err := fileops.ListFilesByPrefixAndSuffix(filepath.Join(workDir, "bin"), "", "")
		if err != nil {
			return err
		}
		binaries := []string{}
		for _, file := range files {
			binaries = append(binaries, filepath.Base(file))
		}
		// binaries := []string{"scion", "control", "router", "dispatcher", "gateway", "daemon"}
		for _, binary := range binaries {
			log.Println("[Install] Copy binary ", filepath.Join(workDir, "bin", binary), "to ", filepath.Join(binPath, binary))
			err = fileops.CopyFile(filepath.Join(binPath, binary), filepath.Join(workDir, "bin", binary))
			if err != nil {
				return err
			}
			err = os.Chmod(filepath.Join(binPath, binary), 0777)
			if err != nil {
				return err
			}
		}
		// return installMacService(serviceName, binaryPath, configPath)
	case "windows":
		// TODO: Since there is no bin folder in Windows, copy the files to a folder and add it to PATH
		// https://stackoverflow.com/questions/44272416/how-to-add-a-folder-to-path-environment-variable-in-windows-10-with-screensho
		return fmt.Errorf("Windows is not supported yet")
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

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	localConfigFolder := filepath.Join(wd, "config")
	destSciondFile := filepath.Join(endhostEnv.ConfigPath, "sciond.toml")
	destScionASFile := filepath.Join(endhostEnv.ConfigPath, "scion-orchestrator.toml")

	localBorderRouterConfigFiles, err := fileops.ListFilesByPrefixAndSuffix(localConfigFolder, "br-", ".toml")
	if err != nil {
		return err
	}
	localControlConfigs, err := fileops.ListFilesByPrefixAndSuffix(localConfigFolder, "cs-", ".toml")
	if err != nil {
		return err
	}

	log.Println("[Install] Copying config folder from ", localConfigFolder, " to ", endhostEnv.ConfigPath)
	err = fileops.CopyDir(localConfigFolder, endhostEnv.ConfigPath)
	if err != nil {
		return err
	}

	files := []string{destSciondFile, destScionASFile}

	for _, file := range localBorderRouterConfigFiles {
		correctFile := strings.ReplaceAll(file, fileops.AppendPathSeperatorIfMissing(localConfigFolder), endhostEnv.ConfigPath)
		files = append(files, correctFile)
	}

	for _, file := range localControlConfigs {
		correctFile := strings.ReplaceAll(file, fileops.AppendPathSeperatorIfMissing(localConfigFolder), endhostEnv.ConfigPath)
		files = append(files, correctFile)
	}

	for _, file := range files {
		log.Println("[Install] Configuring ", file, "...")
		err = fileops.ReplaceStringInFile(file, "{configDir}", endhostEnv.ConfigPath)
		if err != nil {
			return errors.New("Failed to configure configDir in " + file + ": " + err.Error())
		}

		err = fileops.ReplaceStringInFile(file, "{databaseDir}", endhostEnv.DatabasePath)
		if err != nil {
			return errors.New("Failed to configure databaseDir config in " + file + ": " + err.Error())
		}

		err = fileops.ReplaceSingleBackslashWithDouble(file)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("[Install] Configured ", file)
	}
	return nil

}
