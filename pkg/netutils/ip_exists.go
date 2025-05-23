package netutils

import (
	"fmt"
	"net"
)

// IsLocalIPWithMTU checks if the given IP exists on a local interface
// and whether the provided MTU is less than or equal to that interface's MTU.
func IsLocalIPWithMTU(ipStr string, mtu int) (bool, error) {
	inputIP := net.ParseIP(ipStr)
	if inputIP == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return false, fmt.Errorf("error getting interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip != nil && ip.Equal(inputIP) {
				if mtu <= iface.MTU {
					return true, nil
				}
				return false, fmt.Errorf("provided MTU (%d) exceeds interface MTU (%d) for IP %s", mtu, iface.MTU, ipStr)
			}
		}
	}

	return false, fmt.Errorf("IP address %s not found on this host", ipStr)
}
