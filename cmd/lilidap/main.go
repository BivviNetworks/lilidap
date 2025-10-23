package main

import (
	"flag"
	"lilidap/internal/ldapserver"
	"log"
)

func main() {
	var listenAddr string

	flag.StringVar(&listenAddr, "listen", ":389", "Address to listen on for incoming LDAP connections")
	flag.Parse()

	// TODO: Add flag for SSH public key or read from config file
	// For now, this is a placeholder that won't actually work
	server := ldapserver.NewServer(listenAddr, nil)

	log.Printf("Starting LDAP server on %s", listenAddr)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
