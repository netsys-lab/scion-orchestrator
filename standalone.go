package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/netsys-lab/scion-as/conf"
	"github.com/netsys-lab/scion-as/environment"
	"github.com/netsys-lab/scion-as/pkg/fileops"
)

func runStandaloneRouter(env environment.EndhostEnvironment, service conf.SCIONService) error {

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

func runStandaloneControlService(env environment.EndhostEnvironment, service conf.SCIONService) error {

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
