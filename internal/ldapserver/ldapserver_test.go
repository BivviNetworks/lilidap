package ldapserver

import (
	"fmt"
	"lilidap/internal/derived"
	"lilidap/internal/testutils/ssh_helpers"
	"lilidap/internal/testutils/tcp_helpers"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestLDAPServer(t *testing.T) {
	assert := assert.New(t)

	// Generate a test SSH key
	_, pubKey, _, err := ssh_helpers.GenerateKeys(1024)
	if err != nil {
		t.Fatal(err)
	}

	// Get a free port for the LDAP server
	port, err := tcp_helpers.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	// Create a temporary LDAP server for testing
	server := NewServer(fmt.Sprintf("localhost:%d", port), pubKey)

	// Start the server in a goroutine
	go func() {
		err := server.Start()
		if err != nil {
			t.Errorf("Failed to start LDAP server: %v", err)
		}
	}()
	defer server.Stop()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual port the server is listening on
	addr := server.Addr()
	t.Logf("LDAP server address: %s", addr)

	// Test basic connectivity
	t.Run("Basic connectivity", func(t *testing.T) {
		conn, err := ldap.Dial("tcp", addr)
		assert.NoError(err, "Should connect to LDAP server")
		if conn != nil {
			defer conn.Close()
			t.Log("Successfully connected to LDAP server")
		}
	})

	// Test LDAP bind failures
	t.Run("Bind failures", func(t *testing.T) {
		testCases := []struct {
			name     string
			dn       string
			password string
			errMsg   string
		}{
			{
				name:     "Invalid DN format",
				dn:       "uid=test,ou=users,dc=example,dc=com",
				password: "localhost:22",
				errMsg:   "DN must start with cn=",
			},
			{
				name:     "Invalid password format",
				dn:       fmt.Sprintf("cn=%s,ou=campers,dc=0_1_0,dc=bivvi", string(ssh.MarshalAuthorizedKey(pubKey))),
				password: "not-a-hostport",
				errMsg:   "Invalid host:port format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				conn, err := ldap.Dial("tcp", addr)
				assert.NoError(err, "Should connect to LDAP server")
				defer conn.Close()

				_, err = conn.SimpleBind(&ldap.SimpleBindRequest{
					Username: tc.dn,
					Password: tc.password,
				})
				assert.Error(err, "Should fail to bind")
				if err != nil {
					t.Logf("Got expected error: %v", err)
				}
			})
		}
	})

	// Test LDAP bind with SSH server validation (success case)
	t.Run("Bind success with SSH validation", func(t *testing.T) {
		// Start a mock SSH server
		ssh_helpers.WithSSHServer(t, 1024, &ssh.ServerConfig{
			PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
				return &ssh.Permissions{}, nil
			},
		}, func(sshPubKey ssh.PublicKey, sshPort int) {
			// Create DN with the SSH public key
			keyBytes := ssh.MarshalAuthorizedKey(sshPubKey)
			dn := fmt.Sprintf("cn=%s,ou=campers,dc=0_1_0,dc=bivvi", string(keyBytes[:len(keyBytes)-1])) // Remove trailing newline

			// Connect to LDAP server
			conn, err := ldap.Dial("tcp", addr)
			assert.NoError(err, "Should connect to LDAP server")
			defer conn.Close()

			// Try to bind with correct host:port (use 127.0.0.1 not localhost since that's the client IP)
			_, err = conn.SimpleBind(&ldap.SimpleBindRequest{
				Username: dn,
				Password: fmt.Sprintf("127.0.0.1:%d", sshPort),
			})
			assert.NoError(err, "Should bind successfully with valid SSH server")
		})
	})

	// Test LDAP search for user attributes
	t.Run("Search for user attributes", func(t *testing.T) {
		// Create DN with the SSH public key (remove trailing newline and trim any whitespace)
		keyBytes := ssh.MarshalAuthorizedKey(pubKey)
		keyStr := string(keyBytes)
		keyStr = keyStr[:len(keyStr)-1] // Remove trailing newline
		dn := fmt.Sprintf("cn=%s,ou=campers,dc=0_1_0,dc=bivvi", keyStr)

		// Connect to the LDAP server
		conn, err := ldap.Dial("tcp", addr)
		assert.NoError(err, "Should connect to LDAP server")
		defer conn.Close()

		// Get the expected attributes from the test key
		attrs := derived.FromPublicKey(pubKey)
		username := attrs.Username()
		displayName := attrs.DisplayName("en")
		phoneNumber := attrs.PhoneNumber()

		// Search for the user
		searchRequest := ldap.NewSearchRequest(
			dn,
			ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
			"(objectClass=*)",
			[]string{"uid", "displayName", "telephoneNumber", "uidNumber", "gidNumber", "homeDirectory"},
			nil,
		)

		result, err := conn.Search(searchRequest)
		assert.NoError(err, "Should search successfully")
		assert.Equal(1, len(result.Entries), "Should find exactly one user")

		entry := result.Entries[0]
		assert.Equal(username, entry.GetAttributeValue("uid"), "Should have correct username")
		assert.Equal(displayName, entry.GetAttributeValue("displayName"), "Should have correct display name")
		assert.Equal(phoneNumber, entry.GetAttributeValue("telephoneNumber"), "Should have correct phone number")
		assert.NotEmpty(entry.GetAttributeValue("uidNumber"), "Should have uidNumber")
		assert.Equal("1001", entry.GetAttributeValue("gidNumber"), "Should have correct gidNumber")
		assert.Equal(fmt.Sprintf("/home/%s", username), entry.GetAttributeValue("homeDirectory"), "Should have correct homeDirectory")
	})
}
