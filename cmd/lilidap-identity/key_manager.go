package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// getOrCreateKey loads existing key or generates new Ed25519 key at keyPath
// Always saves newly generated keys to disk
func getOrCreateKey(keyPath string) (ssh.Signer, ssh.PublicKey, error) {
	// Check if key file exists
	if _, err := os.Stat(keyPath); err == nil {
		// Key exists, load it
		return loadPrivateKey(keyPath)
	} else if !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("error checking key file: %w", err)
	}

	// Key doesn't exist, generate new one
	fmt.Printf("🔑 Generating new Ed25519 key at %s\n", keyPath)

	privateKey, signer, err := generateEd25519Key()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Save the key
	if err := saveEd25519Key(privateKey, keyPath); err != nil {
		return nil, nil, fmt.Errorf("failed to save key: %w", err)
	}

	fmt.Printf("✅ Key saved successfully\n\n")
	return signer, signer.PublicKey(), nil
}

// loadPrivateKey loads an SSH private key from disk
// Supports Ed25519, RSA, ECDSA formats
func loadPrivateKey(path string) (ssh.Signer, ssh.PublicKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read key file: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return signer, signer.PublicKey(), nil
}

// generateEd25519Key creates a new Ed25519 key pair
// Returns both the raw private key and the signer
func generateEd25519Key() (ed25519.PrivateKey, ssh.Signer, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create signer: %w", err)
	}

	return privateKey, signer, nil
}

// saveEd25519Key writes an Ed25519 private key to disk (OpenSSH format, 0600 permissions)
// Also writes public key to path.pub (0644 permissions)
func saveEd25519Key(privateKey ed25519.PrivateKey, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to OpenSSH format
	privateKeyPEM, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Write private key with restrictive permissions
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)
	if err := os.WriteFile(path, privateKeyBytes, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Create signer to get public key
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer for public key: %w", err)
	}

	// Write public key
	pubKeyPath := path + ".pub"
	pubKeyBytes := ssh.MarshalAuthorizedKey(signer.PublicKey())
	if err := os.WriteFile(pubKeyPath, pubKeyBytes, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

// expandPath expands ~ to home directory
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if path == "~" {
		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:]), nil
	}

	return path, nil
}
