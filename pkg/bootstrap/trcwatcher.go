package bootstrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"time"

	"github.com/netsys-lab/scion-orchestrator/conf"
)

// Structure to hold TRC information
type TRCInfo struct {
	ID struct {
		ISD          int `json:"isd"`
		BaseNumber   int `json:"base_number"`
		SerialNumber int `json:"serial_number"`
	} `json:"id"`
}

// Flag to control the main loop
var running = true

func parseFilename(filename string) *TRCInfo {
	// Use regular expression to extract isd, base, and serial
	re := regexp.MustCompile(`ISD(\d+)-B(\d+)-S(\d+)\.trc`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) == 4 {
		isd, base, serial := matches[1], matches[2], matches[3]

		trcInfo := &TRCInfo{}
		fmt.Sscanf(isd, "%d", &trcInfo.ID.ISD)
		fmt.Sscanf(base, "%d", &trcInfo.ID.BaseNumber)
		fmt.Sscanf(serial, "%d", &trcInfo.ID.SerialNumber)

		return trcInfo
	}
	return nil
}

func createTRCsJSON(directory string, targetFile string) error {
	var trcs []TRCInfo

	// List all files in the directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	// Parse each filename
	for _, file := range files {
		if !file.IsDir() {
			trcInfo := parseFilename(file.Name())
			if trcInfo != nil {
				trcs = append(trcs, *trcInfo)
			}
		}
	}

	// Write the JSON array to a file

	jsonData, err := json.MarshalIndent(trcs, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(targetFile, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func RunTrcFileWatcher(configDir string, config *conf.Config) error {
	// Directory containing the TRC files
	directoryPath := filepath.Join(configDir, "certs")

	// Time interval (in seconds)
	interval := 1 * time.Minute
	log.Println("[Bootstrap Server] Updating trcs.json...")
	for running {

		err := createTRCsJSON(directoryPath, filepath.Join(configDir, "trcs.json"))
		if err != nil {
			return errors.New(fmt.Sprintf("[Bootstrap Server] Error creating TRCs JSON: %v\n", err))
		} else {
			// log.Println("[Bootstrap Server] Update complete. Waiting for next interval...")
		}

		time.Sleep(interval)
	}

	return nil

}
