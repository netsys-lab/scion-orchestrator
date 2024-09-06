package environment

import (
	"fmt"
	"os"
	"path/filepath"
)

func init() {

	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		fmt.Println("The Program Files directory could not be found.")
		os.Exit(1)
	}

	basePath := filepath.Join(programFiles, "scion")
	configPath := basePath

	HostEnv = &HostEnvironment{

		ConfigPath:        configPath,
		BasePath:          basePath,
		DaemonConfigPath:  configPath,
		ControlConfigPath: configPath,
		RouterConfigPath:  configPath,
		DatabasePath:      filepath.Join(basePath, "database"),
		LogPath:           filepath.Join(basePath, "logs"),
	}
}
