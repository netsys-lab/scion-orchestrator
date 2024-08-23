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
	Index      int
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
			Name:       fmt.Sprintf("scion-control-service-cs%d", index+1),
			ConfigFile: controlConfigFile,
			Type:       "control",
			Index:      index + 1,
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
			Name:       fmt.Sprintf("scion-border-router-br%d", index+1),
			ConfigFile: routerConfigFile,
			Type:       "router",
			Index:      index + 1,
		}
		c.BorderRouters = append(c.BorderRouters, service)
	}

	if fileops.FileOrFolderExists(filepath.Join(c.Folder, "dispatcher.toml")) {
		c.Dispatcher = &SCIONService{
			Name:       "scion-dispatcher",
			ConfigFile: filepath.Join(c.Folder, "dispatcher.toml"),
			Type:       "dispatcher",
		}
	}

	if fileops.FileOrFolderExists(filepath.Join(c.Folder, "sciond.toml")) {
		c.Daemon = SCIONService{
			Name:       "scion-daemon",
			ConfigFile: filepath.Join(c.Folder, "sciond.toml"),
			Type:       "daemon",
		}
	}

	return c, nil
}

func (c *SCIONConfig) Log() string {
	return fmt.Sprintf("[SCIONConfig] Got Control Services: %d; Border Routers: %d; Dispatcher: %t; Daemon: %t;",
		len(c.ControlServices), len(c.BorderRouters), c.Dispatcher != nil, true)
}
