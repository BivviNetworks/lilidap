package main

import (
	"fmt"
	"net"
)

// getPort returns an available port
// If requestedPort is 0, finds a free port automatically
// If requestedPort is specified, validates it's available
func getPort(host string, requestedPort int) (int, error) {
	if requestedPort == 0 {
		// Auto-select a free port
		return getFreePort(host)
	}

	// Validate the requested port is available
	addr := fmt.Sprintf("%s:%d", host, requestedPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("port %d is not available: %w", requestedPort, err)
	}
	listener.Close()

	return requestedPort, nil
}

// getFreePort finds and returns a free port on the specified host
func getFreePort(host string) (int, error) {
	addr := fmt.Sprintf("%s:0", host)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("failed to find free port: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	return port, nil
}
