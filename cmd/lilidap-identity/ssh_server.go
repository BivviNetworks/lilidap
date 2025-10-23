package main

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// startSSHServer starts an SSH server and returns a stop function
// Based on StartMockSSHServer from internal/testutils/ssh_helpers
func startSSHServer(signer ssh.Signer, host string, port int) (func(), error) {
	// Create minimal SSH server config
	config := &ssh.ServerConfig{
		NoClientAuth: false,
		MaxAuthTries: 1,
		// Reject all password authentication
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, fmt.Errorf("password authentication not supported")
		},
	}

	// Add the identity key as the server's host key
	config.AddHostKey(signer)

	// Start listening
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Start accepting connections in a goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Listener closed, exit gracefully
				return
			}

			// Handle SSH connection in a goroutine
			go handleSSHConnection(conn, config)
		}
	}()

	// Return stop function
	stopFunc := func() {
		listener.Close()
	}

	return stopFunc, nil
}

// handleSSHConnection processes an SSH connection
func handleSSHConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	// Perform SSH handshake
	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		// Connection failed during handshake (expected - we reject auth)
		return
	}

	// Discard global requests
	go ssh.DiscardRequests(reqs)

	// Reject all channels
	for newChan := range chans {
		if err := newChan.Reject(ssh.Prohibited, "no shell access provided"); err != nil {
			log.Printf("Failed to reject channel: %v", err)
		}
	}
}
