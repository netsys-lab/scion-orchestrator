package scionutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

// SCIONTopology represents the topology format in JSON
type SCIONTopology struct {
	Attributes       []string                   `json:"attributes"`
	ISD_AS           string                     `json:"isd_as"`
	MTU              int                        `json:"mtu"`
	ControlService   map[string]ServiceEndpoint `json:"control_service"`
	DiscoveryService map[string]ServiceEndpoint `json:"discovery_service"`
	BorderRouters    map[string]BorderRouter    `json:"border_routers"`
	DispatchedPorts  string                     `json:"dispatched_ports"`
}

// ServiceEndpoint represents a generic SCION service (e.g., control, discovery)
type ServiceEndpoint struct {
	Addr string `json:"addr"`
}

// BorderRouter represents a SCION border router
type BorderRouter struct {
	InternalAddr string                     `json:"internal_addr"`
	Interfaces   map[string]RouterInterface `json:"interfaces"`
}

// RouterInterface represents a SCION border router interface
type RouterInterface struct {
	Underlay Underlay `json:"underlay"`
	ISD_AS   string   `json:"isd_as"`
	LinkTo   string   `json:"link_to"`
	MTU      int      `json:"mtu"`
}

// Underlay represents the underlay transport information
type Underlay struct {
	Public string `json:"public"`
	Remote string `json:"remote"`
}

// NextInterfaceID iterates all interface IDs and returns the next available one (+1)
func (t *SCIONTopology) NextInterfaceID() int {
	existingIDs := []int{}

	// Collect all interface IDs
	for _, router := range t.BorderRouters {
		for ifaceID := range router.Interfaces {
			id, err := strconv.Atoi(ifaceID)
			if err == nil {
				existingIDs = append(existingIDs, id)
			}
		}
	}

	// If no interfaces exist, start at 1
	if len(existingIDs) == 0 {
		return 1
	}

	// Sort the IDs and return the next available one
	sort.Ints(existingIDs)
	return existingIDs[len(existingIDs)-1] + 1
}

// SaveSCIONTopology saves a SCIONTopology object to a JSON file, creating a backup before overwriting
func SaveSCIONTopology(filename string, topology *SCIONTopology) error {
	// Create a backup if the file already exists
	if _, err := os.Stat(filename); err == nil {
		backupFilename := fmt.Sprintf("%s.bak", filename)
		err := os.Rename(filename, backupFilename)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("Backup created: %s\n", backupFilename)
	}

	// Marshal topology to JSON
	data, err := json.MarshalIndent(topology, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal topology: %w", err)
	}

	// Write to file
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save topology: %w", err)
	}

	fmt.Println("Topology saved successfully.")
	return nil
}

// LoadSCIONTopology reads a topology JSON file and unmarshals it into SCIONTopology struct
func LoadSCIONTopology(filename string) (*SCIONTopology, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var topology SCIONTopology
	if err := json.Unmarshal(data, &topology); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &topology, nil
}
