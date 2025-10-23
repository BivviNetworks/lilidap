package main

import (
	"flag"
	"fmt"
	"lilidap/internal/ldapserver"
	"log"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "", "IP address to bind to (default: all interfaces)")
	flag.IntVar(&port, "port", 389, "Port to listen on")
	flag.Parse()

	// Construct listen address
	listenAddr := fmt.Sprintf("%s:%d", host, port)

	// Note: The sshPubKey parameter is currently unused by the server
	// The server validates SSH keys provided by clients in the BIND DN
	// TODO: When we implement the identity manager (separate SSH server),
	// we'll use this to pass the server's own key for validation
	server := ldapserver.NewServer(listenAddr, nil)

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              LiliDAP LDAP Server Starting                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Display bind address
	bindHost := host
	if bindHost == "" {
		bindHost = "0.0.0.0 (all interfaces)"
	}
	fmt.Printf("🌐 Bind Address: %s\n", bindHost)
	fmt.Printf("🔌 Port: %d\n", port)
	fmt.Printf("📡 Listening on: %s\n", listenAddr)

	// Warnings for privileged ports and defaults
	if port == 389 {
		fmt.Println()
		fmt.Println("   ⚠️  Port 389 requires root/administrator privileges")
		fmt.Println("   💡 For testing without root: use --port 3389")
	}
	if host == "" {
		fmt.Println("   ℹ️  Accessible from any network interface")
		fmt.Println("   💡 For localhost only: use --host localhost")
	}

	fmt.Println()
	fmt.Println("📋 How to authenticate:")
	fmt.Println("   DN (Username): cn=<your-ssh-public-key>,ou=campers,dc=0_1_0,dc=bivvi")
	fmt.Println("   Password: <your-ssh-server-host>:<port>")
	fmt.Println()
	fmt.Println("📝 Example:")
	fmt.Println("   DN: cn=ssh-rsa AAAAB3NzaC1yc2E...,ou=campers,dc=0_1_0,dc=bivvi")
	fmt.Println("   Password: 192.168.1.100:22")
	fmt.Println()
	fmt.Println("🔍 Test with:")

	// Smart display address for test command
	displayAddr := listenAddr
	if host == "" {
		displayAddr = fmt.Sprintf("localhost:%d", port)
	}
	fmt.Printf("   ldapwhoami -H ldap://%s -D \"<dn>\" -w \"<host:port>\"\n", displayAddr)
	fmt.Println()
	fmt.Println("⚠️  Note: Your SSH server must be running and accessible")
	fmt.Println("    from the LDAP server for authentication to work.")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Server running. Press Ctrl+C to stop.")
	fmt.Println()

	if err := server.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
