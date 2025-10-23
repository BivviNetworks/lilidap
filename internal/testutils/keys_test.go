package testutils

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func doOneKey(t *testing.T, name string, keyFunc func() (ssh.PublicKey, error)) {
	key, err := keyFunc()
	if err != nil {
		t.Fatalf("Failed to parse %v test key: %v", name, err)
	}
	if key == nil {
		t.Errorf("Expected non-nil key from %v", name)
	}
}

func TestGetTestPublicKey(t *testing.T) {
	doOneKey(t, "GetTestPublicKey", GetTestPublicKey)
}

func TestGetRandomPublicKey(t *testing.T) {
	doOneKey(t, "GetRandomPublicKey", GetRandomPublicKey)
}
