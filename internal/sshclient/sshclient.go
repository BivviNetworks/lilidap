package sshclient

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

const nonMatchingKeyFlag = "lilidap: server public key did not match"

func ValidateServerPublicKey(serverAddress string, serverPort int, expectedPublicKey ssh.PublicKey, onDebugMessage func(string)) (bool, error) {
	log := func(line string) { onDebugMessage(fmt.Sprintf("ValidateServerPublicKey: %s", line)) }

	clientConfig := &ssh.ClientConfig{
		// User: "dummy", // It doesn't matter in our case, as we're only verifying the key.
		// Auth: []ssh.AuthMethod{
		// 	ssh.Password("dummy"), // Similarly, this doesn't matter.
		// },
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: func(hostname string, remote net.Addr, actualPublicKey ssh.PublicKey) error {
			if bytes.Equal(ssh.MarshalAuthorizedKey(expectedPublicKey), ssh.MarshalAuthorizedKey(actualPublicKey)) {
				log("SSH client HostKeyCallback: correct key presented, proceeding")
				return nil
			}
			log("SSH client HostKeyCallback: wrong key presented; early exit")
			return fmt.Errorf(nonMatchingKeyFlag)
		},
	}

	log("SSH client will now ssh.Dial")
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", serverAddress, serverPort), clientConfig)
	if err != nil {
		log(fmt.Sprintf("SSH client ssh.Dial caught error: %s", err.Error()))
		// If the error is related to key mismatch, return false and no error.
		if strings.Contains(err.Error(), nonMatchingKeyFlag) {
			return false, nil
		}

		// Error: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none], no supported methods remain
		if strings.Contains(err.Error(), "ssh: unable to authenticate") {
			// This is our specific error indicating that the server's key verification passed
			// but user authentication failed as expected.
			return true, nil
		}

		if strings.Contains(err.Error(), "ssh: handshake failed") {
			// these seem to come up interchangeably when the server doesn't accept auth
			if strings.Contains(err.Error(), "ssh: handshake failed: read tcp") && strings.Contains(err.Error(), "read: connection reset by peer") {
				return false, fmt.Errorf("server may not be accepting auth methods")
			}
			if strings.Contains(err.Error(), "ssh: handshake failed: EOF") {
				return false, fmt.Errorf("server may not be accepting auth methods")
			}
		}

		// If there's any other error (e.g., server not reachable), return it.
		//  here are others:
		//   ssh: no authentication methods configured but NoClientAuth is also false
		return false, err
	}
	defer client.Close()

	// If no error occurred in the HostKeyCallback, we did something wrong
	return false, fmt.Errorf("connection succeeded illegally")
}
