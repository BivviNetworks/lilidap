package ldap_helpers

import (
	"fmt"
	"lilidap/internal/derived"
	"lilidap/internal/testutils/ssh_helpers"
	"os"
	"os/exec"
	"testing"
	"time"

	"lilidap/internal/testutils/map_helpers"
	"lilidap/internal/testutils/tcp_helpers"

	"github.com/go-ldap/ldap/v3"
	"gopkg.in/yaml.v3"
)

// TODO: native struct for yaml config
func WithGLauth(t *testing.T, baseConfig map[string]interface{}, body func(int)) {

	port, err := tcp_helpers.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using port: %d", port)

	// Define the configuration using maps.
	// TODO: set an admin user and password, pass them to body()

	dynamicConfig := map[string]interface{}{
		"global": map[string]interface{}{
			"debug": false,
			"port":  port,
		},
		"backend": map[string]interface{}{
			"datastore": "config",
			"baseDN":    "dc=glauth,dc=com",
		},
		"config": map[string]interface{}{
			"admins": []string{"cn=admin,ou=people,dc=glauth,dc=com"}, // Generate dynamically if needed.
		},
	}

	fullConfig := map_helpers.Merge(baseConfig, dynamicConfig)

	// Serialize the map to YAML.
	data, err := yaml.Marshal(fullConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Save to temp file.
	configFile, err := os.CreateTemp("", "glauth")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configFile.Name())

	_, err = configFile.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	// Launch GLauth
	cmd := exec.Command("glauth", "-config", configFile.Name())
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start GLauth: %v", err)
	}
	defer func() {
		err = cmd.Process.Kill()
		if err != nil {
			t.Fatalf("Failed to kill GLauth: %v", err)
		}
	}()

	// Give GLauth some time to start ... TODO wait for port
	time.Sleep(5 * time.Second)

	// Validate GLauth state using an LDAP client
	conn, err := ldap.Dial("tcp", "localhost:your_GLauth_port")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Example: Validate using a Bind request
	err = conn.Bind("adminDN", "adminPassword")
	if err != nil {
		t.Errorf("Expected successful bind, got: %v", err)
	}

	body(port)
}

// ConnectAndBindWithTestKey connects to an LDAP server and binds using the test key
func ConnectAndBindWithTestKey(addr string) (*ldap.Conn, error) {
	// Connect to the LDAP server
	conn, err := ldap.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %v", err)
	}

	// Get the test key
	_, pubKey, _, err := ssh_helpers.GenerateKeys(1024)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get test public key: %v", err)
	}

	// Get the expected username from the test key
	attrs := derived.FromPublicKey(pubKey)
	username := attrs.Username()

	// Bind with the username and signature
	_, err = conn.SimpleBind(&ldap.SimpleBindRequest{
		Username: fmt.Sprintf("uid=%s,ou=users,dc=example,dc=com", username),
		Password: "test",
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to bind: %v", err)
	}

	return conn, nil
}
