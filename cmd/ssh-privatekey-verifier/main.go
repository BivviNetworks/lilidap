package main

import (
	"flag"
	"fmt"
	"lilidap/internal/sshclient"

	"golang.org/x/crypto/ssh"
)

func main() {
	host := flag.String("host", "127.0.0.1", "SSH server host")
	port := flag.Int("port", 22, "SSH server port")
	keyString := flag.String("key", "", "The public key text, e.g. 'ssh-rsa AAAAB3Nz...'")
	debug := flag.Bool("debug", false, "Whether to print debug info aimed at the developer")

	flag.Parse()

	if *keyString == "" {
		fmt.Println("Please provide a valid public key using the -key flag.")
		return
	}

	// Parse the SSH public key string
	sshPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*keyString))
	if err != nil {
		fmt.Println("Error parsing SSH public key:", err)
		return
	}

	// Use sshclient to validate server's key
	// Note: You'll need to implement the ValidateServerKey function or similar in your sshclient package.
	valid, err := sshclient.ValidateServerPublicKey(*host, *port, sshPublicKey, func(msg string) {
		if *debug {
			fmt.Println(msg)
		}
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if valid {
		fmt.Println("SSH server proved ownership of the private key.")
	} else {
		fmt.Println("SSH server failed to prove ownership of the private key.")
	}
}
