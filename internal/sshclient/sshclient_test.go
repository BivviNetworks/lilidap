package sshclient

import (
	"strings"
	"testing"

	"lilidap/internal/testutils/ssh_helpers"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const serverAddress = "localhost"
const privKeyLength = 1024
const pubKeyStringLength = 213

// ensure that we can start an SSH server as expected
func TryOneSSHServer(t *testing.T, config *ssh.ServerConfig) {
	ssh_helpers.WithSSHServer(t, privKeyLength, config, func(pubKey ssh.PublicKey, port int) {
		pubKeyStr := string(ssh.MarshalAuthorizedKey(pubKey)) // e.g. "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAEQCvhlzUS+GjQzqkJcxPa0Hr\n"

		// just check that the key is reasonable
		require.True(t, strings.HasPrefix(pubKeyStr, "ssh-rsa "))
		require.Equal(t, pubKeyStringLength, len(pubKeyStr))
	})
}

// start a server and ensure that a client behaves as expected
func TryOneSSHClient(t *testing.T, mac ssh_helpers.MaybeAcceptableConfig) {

	_, wrongPubKey, _, err := ssh_helpers.GenerateKeys(privKeyLength)
	if err != nil {
		t.Fatal(err)
	}

	// work function that applies the expected output to the actual output
	doValidation := func(port int, keyLabel string, keyToUse ssh.PublicKey, criteria ssh_helpers.ValidationResponse) {
		actualValidity, err := ValidateServerPublicKey(serverAddress, port, keyToUse, func(msg string) { t.Log(msg) })
		t.Logf("Validing the %s key got result: %t", keyLabel, actualValidity)
		ssh_helpers.EvaluateResponse(t, criteria, actualValidity, err)
	}

	// script up the valid and invalid cases
	ssh_helpers.WithSSHServer(t, privKeyLength, &mac.Config, func(pubKey ssh.PublicKey, port int) {
		doValidation(port, "correct", pubKey, mac.WhenValid)
		doValidation(port, "incorrect", wrongPubKey, mac.WhenInvalid)
	})
}

// iterate over all SSH configs and dynamically run server tests
func TestSSHServers(t *testing.T) {
	for configName, possibleConfig := range ssh_helpers.SampleServerConfigs {
		t.Run(configName, func(t *testing.T) {
			pcc := &possibleConfig.Config
			TryOneSSHServer(t, pcc)
		})
	}
}

// iterate over all SSH configs and dynmaically run client tests
func TestSSHClients(t *testing.T) {
	for configName, possibleConfig := range ssh_helpers.SampleServerConfigs {
		t.Run(configName, func(t *testing.T) {
			pc := possibleConfig
			TryOneSSHClient(t, pc)
		})
	}
}
