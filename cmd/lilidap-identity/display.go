package main

import (
	"fmt"
	"strings"

	"lilidap/internal/derived"

	"golang.org/x/crypto/ssh"
)

// displayBanner shows the startup banner
func displayBanner() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           LiliDAP Identity Manager                            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// displayServerInfo shows SSH server information
func displayServerInfo(host string, port int, keyPath string, keyType string, fingerprint string) {
	fmt.Printf("🌐 SSH Server: Running on %s:%d\n", host, port)
	fmt.Printf("🔐 SSH Key: %s (%s)\n", keyPath, keyType)
	fmt.Printf("🔑 Fingerprint: %s\n", fingerprint)
	fmt.Println()
}

// displayCredentials shows LDAP credentials
func displayCredentials(pubKey ssh.PublicKey, host string, port int) {
	// Normalize the public key
	normalizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubKey)))

	// Construct the DN
	dn := fmt.Sprintf("cn=%s,ou=campers,dc=0_1_0,dc=bivvi", normalizedKey)
	password := fmt.Sprintf("%s:%d", host, port)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📋 COPY THESE CREDENTIALS:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Username (DN):")
	fmt.Println(dn)
	fmt.Println()
	fmt.Println("Password:")
	fmt.Println(password)
	fmt.Println()
}

// displayDerivedIdentity shows derived identity attributes
func displayDerivedIdentity(attrs *derived.UserAttributes) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("👤 YOUR DERIVED IDENTITY:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("POSIX Username:  %-13s (for file systems, IRC, etc.)\n", attrs.Username())
	fmt.Printf("Friendly Name:   %-13s (for display, caller ID)\n", attrs.DisplayName("en"))
	fmt.Printf("Phone Number:    %-13s (for VoIP)\n", attrs.PhoneNumber())
	fmt.Printf("User ID:         %-13d (numeric UID)\n", attrs.PosixUserID())
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// displayReadyMessage shows the final ready message
func displayReadyMessage() {
	fmt.Println("✅ Ready to authenticate! Use these credentials with any LDAP-enabled service.")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()
}
