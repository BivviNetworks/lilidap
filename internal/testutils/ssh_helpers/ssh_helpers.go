package ssh_helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"lilidap/internal/testutils/tcp_helpers"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

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

// convert a string to a strongly typed key
func PublicKeyOfString(keyString string) (ssh.PublicKey, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyString))
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

// different SSH server configs we will try, and how we expect
//
//	ValidateServerPublicKey to behave with both valid and invalid keys
var SampleServerConfigs = map[string]MaybeAcceptableConfig{
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

func GenerateKeys(keyLength int) (ssh.Signer, ssh.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil {
		return nil, nil, err
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	return signer, signer.PublicKey(), nil
}

// return a function for stopping the server
func StartMockSSHServer(t *testing.T, wg *sync.WaitGroup, config *ssh.ServerConfig, port int, privateKey ssh.Signer) (func(), error) {
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
				conn.Close()
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

	return func() { listener.Close() }, nil
}

func WithSSHServer(t *testing.T, keyLength int, config *ssh.ServerConfig, body func(ssh.PublicKey, int)) {
	t.Log("WithServer begins")
	privateKey, pubKey, err := GenerateKeys(keyLength)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using pubKey: %s", ssh.MarshalAuthorizedKey(pubKey))

	port, err := tcp_helpers.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using port: %d", port)

	t.Log("Initializing WaitGroup")
	var wg sync.WaitGroup
	defer wg.Wait()

	t.Log("Starting server")
	stopServer, err := StartMockSSHServer(t, &wg, config, port, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	defer stopServer()

	// when the port opens, turn it over to the work function
	tcp_helpers.WaitForPort(t, "localhost", port)
	time.Sleep(50 * time.Millisecond) // give the server a chance to close the connection; makes logs nicer
	t.Log("Port open, starting WithServer body function")
	body(pubKey, port)
}

func EvaluateResponse(t *testing.T, expected ValidationResponse, actualValidity bool, actualErr error) {
	if expected.HasError {
		require.NotNil(t, actualErr)
		t.Logf("We expected an error and got one.  Now checking if '%s' = '%s'", expected.ErrorMsg, actualErr.Error())
		require.Equal(t, expected.ErrorMsg, actualErr.Error())
	} else {
		if actualErr != nil {
			t.Fatal(actualErr)
		}
		t.Logf("We expected no error and didn't get one.  Now checking if %t = %t", expected.Success, actualValidity)
		require.Equal(t, expected.Success, actualValidity)
	}
}
