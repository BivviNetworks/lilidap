package sshclient

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"golang.org/x/crypto/ssh"
)

const serverAddress = "localhost"

func generateKeys() (ssh.Signer, ssh.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	return signer, signer.PublicKey(), nil
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", serverAddress))
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func waitForPort(t *testing.T, port int) {
	t.Log("waitForPort begins")
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverAddress, port))
		if err == nil {
			t.Log("waitForPort success")
			conn.Close()
			break
		}
		t.Logf("waitForPort must keep waiting: %v", err)
		time.Sleep(100 * time.Millisecond)
	}
}

func startMockSSHServer(t *testing.T, privateKey ssh.Signer, port int) (net.Listener, error) {
	log := func(line string) { t.Logf("StartMockSSHServer: %s", line) }

	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	config.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log(fmt.Sprintf("net.Listen fail: %v", err))
		return nil, err
	}
	log("net.Listen-ing")

	go func() {
		log := func(line string) { t.Logf("MockSSHServer: %s", line) }
		for {
			log("listening")
			conn, err := listener.Accept()
			if err != nil {
				log(fmt.Sprintf("listener.Accept fail (probably quitting time): %v", err))
				return
			}
			log("accepted a connection")

			_, chans, reqs, err := ssh.NewServerConn(conn, config)
			if err != nil {
				log(fmt.Sprintf("NewServerConn fail: %v", err))
				conn.Close() // add a "continue" here if this is switched to the infinite loop version
				log("Closed connection")
				continue
			}

			// Discard global requests
			go ssh.DiscardRequests(reqs)

			// Discard new channels
			for newChan := range chans {
				log("Mock SSH server rejecting a chan")
				if err := newChan.Reject(ssh.Prohibited, "operation not supported"); err != nil {
					log(fmt.Sprintf("Failed to reject channel: %s", err))
				}
			}
		}
	}()
	time.Sleep(50 * time.Millisecond) // give the thread a chance to start; makes logs nicer

	log("listener accepting connections")

	return listener, nil
}

func WithServer(t *testing.T, body func(ssh.PublicKey, int)) {
	t.Log("WithServer begins")
	privateKey, pubKey, err := generateKeys()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using pubKey: %s", ssh.MarshalAuthorizedKey(pubKey))

	port, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using port: %d", port)

	t.Log("Starting server")
	listener, err := startMockSSHServer(t, privateKey, port)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// when the port opens, turn it over to the work function
	waitForPort(t, port)
	time.Sleep(50 * time.Millisecond) // give the server a chance to close the connection; makes logs nicer
	t.Log("Port open, starting WithServer body function")
	body(pubKey, port)
}

func TestSSHServer(t *testing.T) {
	WithServer(t, func(pubKey ssh.PublicKey, port int) {
		pubKeyStr := string(ssh.MarshalAuthorizedKey(pubKey)) // e.g. "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAEQCvhlzUS+GjQzqkJcxPa0Hr\n"

		// just check that the key is reasonable
		require.True(t, strings.HasPrefix(pubKeyStr, "ssh-rsa "))
		require.Equal(t, 213, len(pubKeyStr))
	})
}

func TestSSHClient(t *testing.T) {
	WithServer(t, func(pubKey ssh.PublicKey, port int) {
		t.Log("TestSSHClient: in body")
		isValid, err := ValidateServerPublicKey(serverAddress, port, pubKey, func(msg string) { t.Log(msg) })
		if err != nil {
			t.Fatal(err)
		}
		require.True(t, isValid)

	})
}
