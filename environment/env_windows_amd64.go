package environment

import (
	"embed"
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

	EndhostEnv = &EndhostEnvironment{

		ConfigPath:       configPath,
		BasePath:         basePath,
		DaemonConfigPath: configPath,
	}
}
