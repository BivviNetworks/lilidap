package ldapserver

import (
	"fmt"
	"lilidap/internal/derived"
	"lilidap/internal/sshclient"
	"net"
	"strconv"
	"strings"

	"github.com/lor00x/goldap/message"
	ldap "github.com/vjeantet/ldapserver"
	"golang.org/x/crypto/ssh"
)

type LDAPServer struct {
	server    *ldap.Server
	sshAddr   string
	sshPubKey ssh.PublicKey
}

func NewServer(sshAddr string, sshPubKey ssh.PublicKey) *LDAPServer {
	server := ldap.NewServer()

	s := &LDAPServer{
		server:    server,
		sshAddr:   sshAddr,
		sshPubKey: sshPubKey,
	}

	// Register handlers for specific LDAP operations
	routes := ldap.NewRouteMux()
	routes.Bind(s.handleBind)
	routes.Search(s.handleSearch)
	server.Handle(routes)

	return s
}

func (s *LDAPServer) handleBind(w ldap.ResponseWriter, m *ldap.Message) {
	bindReq := m.GetBindRequest()

	// Parse host:port from password
	hostPort := bindReq.AuthenticationSimple().String()
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid host:port format: %v", err))
		w.Write(res)
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid port: %v", err))
		w.Write(res)
		return
	}

	if port < 1 || port > 65535 {
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid port: %d", port))
		w.Write(res)
		return
	}

	// Get client address from connection
	clientAddr := m.Client.Addr().String()
	clientHost, _, _ := net.SplitHostPort(clientAddr)

	// Ensure the client host matches the host part of the password
	if clientHost != host {
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage("Client host does not match the host in the password")
		w.Write(res)
		return
	}

	// Extract SSH public key from CN in DN
	// DN format: cn=<base32-username>,ou=campers,dc=0_1_0,dc=bivvi
	dn := string(bindReq.Name())
	cn, err := extractCN(dn)
	if err != nil {
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid DN format: %v", err))
		w.Write(res)
		return
	}

	pubKey, err := ssh.ParsePublicKey([]byte(cn))
	if err != nil {
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid SSH key: %v", err))
		w.Write(res)
		return
	}

	// Validate key against SSH server
	var sshDebugMsg string = ""
	onDebug := func(message string) {
		sshDebugMsg = message
	}

	valid, err := sshclient.ValidateServerPublicKey(host, port, pubKey, onDebug)
	if err != nil || !valid {
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(sshDebugMsg)
		w.Write(res)
		return
	}

	res := ldap.NewBindResponse(ldap.LDAPResultSuccess)
	w.Write(res)
}

func (s *LDAPServer) handleSearch(w ldap.ResponseWriter, m *ldap.Message) {
	searchReq := m.GetSearchRequest()

	// Extract SSH public key from CN in base DN
	dn := string(searchReq.BaseObject())
	cn, err := extractCN(dn)
	if err != nil {
		res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultInvalidDNSyntax)
		w.Write(res)
		return
	}

	pubKey, err := ssh.ParsePublicKey([]byte(cn))
	if err != nil {
		res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultInvalidDNSyntax)
		w.Write(res)
		return
	}

	// Generate derived attributes
	attrs := derived.FromPublicKey(pubKey)

	// Return entry with derived attributes
	e := ldap.NewSearchResultEntry(dn)
	e.AddAttribute("objectClass", message.AttributeValue("inetOrgPerson"))
	e.AddAttribute("uid", message.AttributeValue(attrs.Username()))
	e.AddAttribute("mobile", message.AttributeValue(attrs.PhoneNumber()))
	e.AddAttribute("displayName", message.AttributeValue(attrs.DisplayName("en")))
	w.Write(e)

	res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
	w.Write(res)
}

// extractCN extracts the CN value from a DN string
// DN format: cn=<value>,ou=campers,dc=0_1_0,dc=bivvi
func extractCN(dn string) (string, error) {
	if !strings.HasPrefix(dn, "cn=") {
		return "", fmt.Errorf("DN must start with cn=")
	}
	parts := strings.Split(dn, ",")
	if len(parts) != 4 {
		return "", fmt.Errorf("DN must have exactly 4 parts")
	}
	if parts[1] != "ou=campers" {
		return "", fmt.Errorf("second part must be ou=campers")
	}
	if parts[2] != "dc=0_1_0" {
		return "", fmt.Errorf("third part must be dc=0_1_0")
	}
	if parts[3] != "dc=bivvi" {
		return "", fmt.Errorf("fourth part must be dc=bivvi")
	}
	return strings.TrimPrefix(parts[0], "cn="), nil
}

func (s *LDAPServer) Start() error {
	return s.server.ListenAndServe(":389")
}
