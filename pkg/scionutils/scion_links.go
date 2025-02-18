package scionutils

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Link struct representing a router interface link
type SCIONLink struct {
	Router      string
	InterfaceID string
	Local       string
	Remote      string
	LinkType    string
	Neighbour   string
	MTU         int
	Up          bool
}

func GetSCIONLinksWithStatus(endpoint string, topology *SCIONTopology) ([]SCIONLink, error) {
	metrics, err := FetchMetrics(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	links := ParseRouterInterfaces(metrics, topology)
	return links, nil

}

// FetchMetrics fetches the SCION router metrics from the given endpoint
func FetchMetrics(endpoint string) ([]string, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var metrics []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		metrics = append(metrics, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading metrics response: %w", err)
	}

	return metrics, nil
}

// ParseRouterInterfaces extracts relevant link state information
func ParseRouterInterfaces(metrics []string, topology *SCIONTopology) []SCIONLink {
	var links []SCIONLink
	re := regexp.MustCompile(`router_interface_up\{interface="(\d+)",isd_as="([^"]+)",neighbor_isd_as="([^"]+)"\} (\d+)`)

	for _, line := range metrics {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			interfaceID := matches[1]
			//_localISDAS := matches[2]
			neighborISDAS := matches[3]
			status := matches[4] // "1" = UP, "0" = DOWN
			//log.Println("Matches:", matches)
			//log.Println("Interface:", interfaceID, "Neighbor:", neighborISDAS, "Status:", status)

			// Find corresponding interface in the topology
			for router, br := range topology.BorderRouters {
				if br.Interfaces[interfaceID].ISD_AS == neighborISDAS {
					links = append(links, SCIONLink{
						Router:      router,
						InterfaceID: interfaceID,
						Local:       br.Interfaces[interfaceID].Underlay.Public,
						Remote:      br.Interfaces[interfaceID].Underlay.Remote,
						MTU:         br.Interfaces[interfaceID].MTU,
						Up:          status == "1",
						Neighbour:   neighborISDAS,
						LinkType:    br.Interfaces[interfaceID].LinkTo,
					})
				}
			}
		}
	}

	return links
}

// SearchMetrics filters metrics based on a given search term
func SearchMetrics(metrics []string, searchTerm string) []string {
	var results []string
	for _, line := range metrics {
		if strings.Contains(line, searchTerm) {
			results = append(results, line)
		}
	}
	return results
}
