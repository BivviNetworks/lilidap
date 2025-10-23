package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"lilidap/internal/derived"

	"golang.org/x/crypto/ssh"
)

func main() {
	// Parse flags
	keyPath := flag.String("key", "~/.lilidap/identity", "Path to SSH private key")
	sshHost := flag.String("ssh-host", "127.0.0.1", "SSH server host")
	sshPort := flag.Int("ssh-port", 0, "SSH server port (0=auto)")
	flag.Parse()

	// Show banner
	displayBanner()

	// Expand home directory in key path
	expandedKeyPath, err := expandPath(*keyPath)
	if err != nil {
		log.Fatalf("❌ Invalid key path: %v", err)
	}

	// Load or generate SSH key pair (auto-saves if generated)
	fmt.Println("🔑 Loading identity key...")
	signer, pubKey, err := getOrCreateKey(expandedKeyPath)
	if err != nil {
		log.Fatalf("❌ Key management error: %v", err)
	}

	// Find a free port (or use specified)
	fmt.Println("🔍 Finding available port...")
	port, err := getPort(*sshHost, *sshPort)
	if err != nil {
		log.Fatalf("❌ Port selection error: %v", err)
	}

	// Start SSH server
	fmt.Printf("🚀 Starting SSH server on %s:%d...\n", *sshHost, port)
	stopServer, err := startSSHServer(signer, *sshHost, port)
	if err != nil {
		log.Fatalf("❌ SSH server error: %v", err)
	}
	defer stopServer()
	fmt.Println()

	// Display server info
	keyType := pubKey.Type()
	fingerprint := ssh.FingerprintSHA256(pubKey)
	displayServerInfo(*sshHost, port, expandedKeyPath, keyType, fingerprint)

	// Display LDAP credentials
	displayCredentials(pubKey, *sshHost, port)

	// Derive and display identity attributes
	fmt.Println("🧮 Deriving identity attributes...")
	attrs := derived.FromPublicKey(pubKey)
	displayDerivedIdentity(attrs)

	// Show ready message
	displayReadyMessage()

	// Keep running until interrupted
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println()
	fmt.Println("👋 Shutting down gracefully...")
}
