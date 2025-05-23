package netutils

import (
	"net"
)

// IsUDPPortFree checks if a given UDP address is ready for listen, i.e., if the port is free.
func IsUDPPortFree(addr string) bool {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return false
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

// IsTCPPortFree checks if a given TCP address is ready for listen, i.e., if the port is free.
func IsTCPPortFree(addr string) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return false
	}

	conn, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
