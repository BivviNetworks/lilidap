package sshclient

import (
	"bytes"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
	//"time"
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
			log("SSH client begin HostKeyCallback")
			if bytes.Equal(ssh.MarshalAuthorizedKey(expectedPublicKey), ssh.MarshalAuthorizedKey(actualPublicKey)) {
				log("SSH client HostKeyCallback: keys equal")
				return nil
			}
			log("SSH client HostKeyCallback: keys differ")
			return fmt.Errorf(nonMatchingKeyFlag)
		},
	}

	// onDebugMessage("Will now DialTimeout")
	// conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", serverAddress, serverPort), 5 * time.Second)
	// if err != nil {
	// 	onDebugMessage(fmt.Sprintf("DialTimeout fail: %v", err))
	// 	return false, fmt.Errorf("failed to dial: %v", err)
	// }

	// onDebugMessage("Will now ssh.NewClientConn")
	// sc, chans, reqs, err2 := ssh.NewClientConn(conn, fmt.Sprintf("%s:%d", serverAddress, serverPort), clientConfig)
	// if err != nil {
	// 	onDebugMessage(fmt.Sprintf("NewClientConn fail: %v", err2))
	// 	conn.Close() // Close the underlying connection
	// 	return false, err2
	// }
	// onDebugMessage("Will now ssh.NewClient")
	// client := ssh.NewClient(sc, chans, reqs)
	// onDebugMessage("did ssh.NewClient")
	// defer client.Close() // Ensure the client connection is closed after the validation

	log("SSH client will now ssh.Dial")
	_, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", serverAddress, serverPort), clientConfig)
	if err != nil {
		log(fmt.Sprintf("SSH client ssh.Dial caught error: %s", err.Error()))
		// If the error is related to key mismatch, return false and no error.
		if err.Error() == nonMatchingKeyFlag {
			return false, nil
		}
		// If there's any other error (e.g., server not reachable), return it.
		return false, err
	}

	log("ValidateServerPublicKey success")
	// If no error occurred in the HostKeyCallback, then the server's key is valid.
	return true, nil
}
