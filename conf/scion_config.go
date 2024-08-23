package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/netsys-lab/scion-as/pkg/fileops"
)

type SCIONService struct {
	Name       string
	ConfigFile string
	Type       string
}

type SCIONConfig struct {
	Folder          string
	ControlServices []SCIONService
	BorderRouters   []SCIONService
	Dispatcher      *SCIONService
	Daemon          SCIONService
}

func NewSCIONConfig() *SCIONConfig {
	wd, _ := os.Getwd()
	return &SCIONConfig{
		Folder:          filepath.Join(wd, "config"),
		ControlServices: []SCIONService{},
		BorderRouters:   []SCIONService{},
	}
}

func LoadSCIONConfig() (*SCIONConfig, error) {
	c := NewSCIONConfig()

	// Get all files in c.Folder that start with cs- and end with .toml
	controlConfigFiles, err := fileops.ListFilesByPrefixAndSuffix(c.Folder, "cs-", ".toml")
	if err != nil {
		return nil, err
	}

	for index, controlConfigFile := range controlConfigFiles {
		service := SCIONService{
			Name:       fmt.Sprintf("control-%d", index+1),
			ConfigFile: controlConfigFile,
			Type:       "control",
		}
		c.ControlServices = append(c.ControlServices, service)
	}

	// Get all files in c.Folder that start with cs- and end with .toml
	routerConfigFiles, err := fileops.ListFilesByPrefixAndSuffix(c.Folder, "br-", ".toml")
	if err != nil {
		return nil, err
	}

	for index, routerConfigFile := range routerConfigFiles {
		service := SCIONService{
			Name:       fmt.Sprintf("router-%d", index+1),
			ConfigFile: routerConfigFile,
			Type:       "router",
		}
		c.BorderRouters = append(c.BorderRouters, service)
	}

	if fileops.FileOrFolderExists(filepath.Join(c.Folder, "dispatcher.toml")) {
		c.Dispatcher = &SCIONService{
			Name:       "Dispatcher",
			ConfigFile: "dispatcher.toml",
			Type:       "dispatcher",
		}
	}

	if fileops.FileOrFolderExists(filepath.Join(c.Folder, "sciond.toml")) {
		c.Dispatcher = &SCIONService{
			Name:       "Daemon",
			ConfigFile: "sciond.toml",
			Type:       "daemon",
		}
	}

	return c, nil
}

func (c *SCIONConfig) Log() string {
	return fmt.Sprintf("Control Services: %v\nBorder Routers: %v\nDispatcher: %v\nDaemon: %v\n",
		c.ControlServices, c.BorderRouters, c.Dispatcher, c.Daemon)
}
