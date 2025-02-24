package testutils

import (
	"crypto/rand"
	"crypto/rsa"

	"lilidap/internal/testutils/ssh_helpers"

	"golang.org/x/crypto/ssh"
)

// TestPublicKeyString is in OpenSSH authorized_keys format
const TestPublicKeyString = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCqMQOAqYhVGXxLRjZUVE6cZ6gEYQhYKrRsP0aIBijHWyPGo+ccDwHwsZ5PBhF4UNOGkGgPZt6NHhgl0G4qEGWVtZVhr5dX8NGxwm/ZQYxhj1WV0WldkxGxzb9KG6sQqpD7YZxPkEwVZI2bJA3h0qcOi4/FOY+bL5YAHzTK9QMqrnZcVx3UhGI9h2Gpk2LJJ8xvQPPPbUPHwNzxDuL3UHqPOwQYVixG29NMGXqA4QdDPpH4Poff7hR1sGPxULPKaefhysQ0qz1ezhYQxCjzKIGOwgwYvxgk1JtNp3EKZLtl1B2nwUY9Uu7p44TH/JvJBCkkKiIYbV8Tj8NkH9jskG5v test@lilidap`

func GetRandomPublicKey() (ssh.PublicKey, error) {
	// Generate a 2048-bit RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Convert to SSH format
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, err
	}

	// Marshal to authorized_keys format
	return signer.PublicKey(), nil
}

// GetTestPublicKeyRaw returns the raw key data
func GetTestPublicKeyRaw() ([]byte, error) {
	_, _, _, raw, err := ssh.ParseAuthorizedKey([]byte(TestPublicKeyString))
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// GetTestPublicKey returns a parsed SSH public key for testing
func GetTestPublicKey() (ssh.PublicKey, error) {
	return ssh_helpers.PublicKeyOfString(TestPublicKeyString)
}
