package netutils

import (
	"fmt"
	"net"
)

// Function to check if the source IP is in a blocked subnet
func IsIPInSubnets(ip string, subnets []string) (bool, error) {
	for _, subnet := range subnets {
		_, cidr, err := net.ParseCIDR(subnet)
		if err != nil {
			return false, fmt.Errorf("Invalid CIDR in subnets: %s", subnet)
		}
		parsedIP := net.ParseIP(ip)
		if cidr.Contains(parsedIP) {
			return true, nil
		}
	}
	return false, nil
}
