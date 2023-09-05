package sshclient

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const serverAddress = "localhost"

type ValidationResponse struct {
	HasError bool
	ErrorMsg string
	Success  bool
}

type MaybeAcceptableConfig struct {
	Config      ssh.ServerConfig
	WhenValid   ValidationResponse
	WhenInvalid ValidationResponse
}

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

func startMockSSHServer(t *testing.T, wg *sync.WaitGroup, config *ssh.ServerConfig, port int, privateKey ssh.Signer) (net.Listener, error) {
	log := func(line string) { t.Logf("StartMockSSHServer: %s", line) }

	config.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log(fmt.Sprintf("net.Listen fail: %v", err))
		return nil, err
	}
	log("net.Listen-ing")

	wg.Add(1)
	go func() {
		defer wg.Done()
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

func WithServer(t *testing.T, config *ssh.ServerConfig, body func(ssh.PublicKey, int)) {
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
	var wg sync.WaitGroup
	listener, err := startMockSSHServer(t, &wg, config, port, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	defer wg.Wait()
	defer listener.Close()

	// when the port opens, turn it over to the work function
	waitForPort(t, port)
	time.Sleep(50 * time.Millisecond) // give the server a chance to close the connection; makes logs nicer
	t.Log("Port open, starting WithServer body function")
	body(pubKey, port)
}

func TryOneSSHServer(t *testing.T, config *ssh.ServerConfig) {
	WithServer(t, config, func(pubKey ssh.PublicKey, port int) {
		pubKeyStr := string(ssh.MarshalAuthorizedKey(pubKey)) // e.g. "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAEQCvhlzUS+GjQzqkJcxPa0Hr\n"

		// just check that the key is reasonable
		require.True(t, strings.HasPrefix(pubKeyStr, "ssh-rsa "))
		require.Equal(t, 213, len(pubKeyStr))
	})
}

func TryOneSSHClient(t *testing.T, mac MaybeAcceptableConfig) {

	_, wrongPubKey, err := generateKeys()
	if err != nil {
		t.Fatal(err)
	}

	doValidation := func(port int, expectedValid bool, keyToUse ssh.PublicKey) {
		actualValidity, err := ValidateServerPublicKey(serverAddress, port, keyToUse, func(msg string) { t.Log(msg) })
		var criteria ValidationResponse
		var keyLabel string
		if expectedValid {
			keyLabel = "correct"
			criteria = mac.WhenValid
		} else {
			keyLabel = "incorrect"
			criteria = mac.WhenInvalid
		}
		t.Logf("Validing the %s key got result: %t", keyLabel, actualValidity)

		if criteria.HasError {
			require.NotNil(t, err)
			t.Logf("We expected an error and got one.  Now checking if '%s' = '%s'", criteria.ErrorMsg, err.Error())
			require.Equal(t, criteria.ErrorMsg, err.Error())
		} else {
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("We expected no error and didn't get one.  Now checking if %t = %t", criteria.Success, actualValidity)
			require.Equal(t, criteria.Success, actualValidity)
		}
	}

	WithServer(t, &mac.Config, func(pubKey ssh.PublicKey, port int) {
		doValidation(port, true, pubKey)
		doValidation(port, false, wrongPubKey)
	})
}

var serverConfigs = map[string]MaybeAcceptableConfig{
	// bad server config: they let the client connect without auth
	"NoClientAuth": {
		Config: ssh.ServerConfig{
			NoClientAuth: true,
			MaxAuthTries: 1,
		},
		WhenValid: ValidationResponse{
			HasError: true,
			ErrorMsg: "connection succeeded illegally",
			Success:  false,
		},
		WhenInvalid: ValidationResponse{
			HasError: false,
			ErrorMsg: "",
			Success:  false,
		},
	},
	// bad server config: they require auth but don't accept any methods
	"NoMethods": {
		Config: ssh.ServerConfig{
			NoClientAuth: false,
			MaxAuthTries: 1,
		},
		WhenValid: ValidationResponse{
			HasError: true,
			ErrorMsg: "server may not be accepting auth methods",
			Success:  true,
		},
		WhenInvalid: ValidationResponse{
			HasError: true,
			ErrorMsg: "server may not be accepting auth methods",
			Success:  false,
		},
	},
	// good config: can accept password (which we don't attempt to supply)
	"AuthPassword": {
		Config: ssh.ServerConfig{
			NoClientAuth: false,
			MaxAuthTries: 1,
			PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				return nil, fmt.Errorf("password rejected for %q", c.User()) // Reject all password authentication attempts.
			},
		},
		WhenValid: ValidationResponse{
			HasError: false,
			ErrorMsg: "",
			Success:  true,
		},
		WhenInvalid: ValidationResponse{
			HasError: false,
			ErrorMsg: "",
			Success:  false,
		},
	},
	// good config: can accept public key (which we don't attempt to supply)
	"AuthPublicKey": {
		Config: ssh.ServerConfig{
			NoClientAuth: false,
			MaxAuthTries: 1,
			PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
				return nil, fmt.Errorf("public key rejected for %q", c.User()) // Reject all public key authentication attempts.
			},
		},
		WhenValid: ValidationResponse{
			HasError: false,
			ErrorMsg: "",
			Success:  true,
		},
		WhenInvalid: ValidationResponse{
			HasError: false,
			ErrorMsg: "",
			Success:  false,
		},
	},
}

func TestSSHServers(t *testing.T) {
	for configName, possibleConfig := range serverConfigs {
		t.Run(configName, func(t *testing.T) {
			pcc := &possibleConfig.Config
			TryOneSSHServer(t, pcc)
		})
	}
}

func TestSSHClients(t *testing.T) {
	for configName, possibleConfig := range serverConfigs {
		t.Run(configName, func(t *testing.T) {
			pc := possibleConfig
			TryOneSSHClient(t, pc)
		})
	}
}
