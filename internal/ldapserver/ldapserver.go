package ldapserver

import (
	"fmt"
	"lilidap/internal/derived"
	"lilidap/internal/sshclient"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/lor00x/goldap/message"
	ldap "github.com/vjeantet/ldapserver"
	"golang.org/x/crypto/ssh"
)

// LDAP Server Implementation
//
// DN format: cn=<full-ssh-public-key>,ou=campers,dc=0_1_0,dc=bivvi
// The CN contains the full SSH public key in OpenSSH authorized_keys format
// Example: cn=ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...,ou=campers,dc=0_1_0,dc=bivvi
//
// User Attributes (all derived from SSH key hash):
// - objectClass: inetOrgPerson, posixAccount
// - uid: u1234abcd                 # POSIX username (base32-encoded hash)
// - uidNumber: 753198499           # Numeric UID (from hash)
// - gidNumber: 1001                # Constant group ID
// - homeDirectory: /home/u1234abcd # From uid
// - telephoneNumber: 8753198499    # VoIP number (from hash)
// - displayName: vantumkeirrof     # Friendly name (syllabic)
// - displayName;lang-XX: ...       # Locale-specific variants
// - cn: vantumkeirrof              # Common Name (copy of displayName)
//
// Authentication Flow:
// 1. Client BIND with DN containing full SSH key + password=host:port
// 2. Server validates SSH key ownership by connecting to host:port
// 3. On success, client can SEARCH to get derived attributes
//
// Base32 encoding (for uid only):
// - Character set: o123456789abcdefghikmnpqrstvwxyz
// - Omits confusable characters for visual clarity
// - Applied only to uid attribute, not to DN
//
// Implementation: we use vjeantet/ldapserver as a base

// LDAPServer represents an LDAP server instance
type LDAPServer struct {
	server     *ldap.Server
	sshAddr    string
	sshPubKey  ssh.PublicKey
	listenAddr string
	sessions   sync.Map // Maps client address (string) → bound DN (string)
}

func getKeyInfo(pubKey ssh.PublicKey) (keyType, fingerprint string) {
	return pubKey.Type(), ssh.FingerprintSHA256(pubKey)
}

// NewServer creates a new LDAP server
func NewServer(listenAddr string, sshPubKey ssh.PublicKey) *LDAPServer {
	server := ldap.NewServer()

	s := &LDAPServer{
		server:     server,
		sshAddr:    "localhost:22", // Default SSH server address
		sshPubKey:  sshPubKey,
		listenAddr: listenAddr,
	}

	// Register handlers for specific LDAP operations
	routes := ldap.NewRouteMux()
	routes.Bind(s.handleBind)
	routes.Search(s.handleSearch)
	routes.Extended(s.handleExtended).RequestName("1.3.6.1.4.1.4203.1.11.3")
	server.Handle(routes)

	return s
}

func (s *LDAPServer) handleBind(w ldap.ResponseWriter, m *ldap.Message) {
	bindReq := m.GetBindRequest()
	clientAddr := m.Client.Addr().String()

	log.Printf("🔐 BIND attempt from %s", clientAddr)

	// Always ensure we send a response
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ BIND FAILED: Internal error: %v", r)
			res := ldap.NewBindResponse(ldap.LDAPResultOperationsError)
			res.SetDiagnosticMessage(fmt.Sprintf("Internal error: %v", r))
			w.Write(res)
		}
	}()

	// Parse host:port from password
	hostPort := bindReq.AuthenticationSimple().String()
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		log.Printf("❌ BIND REJECTED: Invalid host:port format: %v", err)
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid host:port format: %v", err))
		w.Write(res)
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("❌ BIND REJECTED: Invalid port: %v", err)
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid port: %v", err))
		w.Write(res)
		return
	}

	if port < 1 || port > 65535 {
		log.Printf("❌ BIND REJECTED: Port out of range: %d", port)
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid port: %d", port))
		w.Write(res)
		return
	}

	// Get client host from connection
	clientHost, _, _ := net.SplitHostPort(clientAddr)

	// Ensure the client host matches the host part of the password
	if clientHost != host {
		log.Printf("❌ BIND REJECTED: Client host %s != password host %s", clientHost, host)
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage("Client host does not match the host in the password")
		w.Write(res)
		return
	}

	// Extract full SSH public key from CN in DN
	// DN format: cn=<full-ssh-key>,ou=campers,dc=0_1_0,dc=bivvi
	// Example: cn=ssh-rsa AAAAB3NzaC1yc2E...,ou=campers,dc=0_1_0,dc=bivvi
	dn := string(bindReq.Name())
	cn, err := extractCN(dn)
	if err != nil {
		log.Printf("❌ BIND REJECTED: Invalid DN format: %v", err)
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid DN format: %v", err))
		w.Write(res)
		return
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(cn))
	if err != nil {
		log.Printf("❌ BIND REJECTED: Invalid SSH key: %v", err)
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(fmt.Sprintf("Invalid SSH key: %v", err))
		w.Write(res)
		return
	}

	// Get key fingerprint for logging
	keyType, fingerprint := getKeyInfo(pubKey)
	log.Printf("   Validating %s key %s against SSH server %s:%d", keyType, fingerprint, host, port)

	// Validate key against SSH server
	var sshDebugMsg string = ""
	onDebug := func(message string) {
		sshDebugMsg = message
	}

	valid, err := sshclient.ValidateServerPublicKey(host, port, pubKey, onDebug)
	if err != nil || !valid {
		log.Printf("❌ BIND REJECTED: SSH validation failed: %s", sshDebugMsg)
		res := ldap.NewBindResponse(ldap.LDAPResultInvalidCredentials)
		res.SetDiagnosticMessage(sshDebugMsg)
		w.Write(res)
		return
	}

	log.Printf("✅ BIND ACCEPTED: %s key %s authenticated successfully", keyType, fingerprint)

	// Normalize the SSH key to ensure consistent representation
	// MarshalAuthorizedKey returns the canonical form with a trailing newline
	normalizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubKey)))

	// Reconstruct the DN with the normalized key
	normalizedDN := fmt.Sprintf("cn=%s,ou=campers,dc=0_1_0,dc=bivvi", normalizedKey)

	// Store the normalized DN in the session for this client
	s.sessions.Store(clientAddr, normalizedDN)

	res := ldap.NewBindResponse(ldap.LDAPResultSuccess)
	w.Write(res)
}

func (s *LDAPServer) handleSearch(w ldap.ResponseWriter, m *ldap.Message) {
	searchReq := m.GetSearchRequest()
	clientAddr := m.Client.Addr().String()

	log.Printf("🔍 SEARCH request from %s", clientAddr)

	// Extract full SSH public key from CN in base DN
	// The CN contains the full SSH key in OpenSSH authorized_keys format
	dn := string(searchReq.BaseObject())
	cn, err := extractCN(dn)
	if err != nil {
		log.Printf("❌ SEARCH REJECTED: Invalid DN format: %v", err)
		res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultInvalidDNSyntax)
		w.Write(res)
		return
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(cn))
	if err != nil {
		log.Printf("❌ SEARCH REJECTED: Invalid SSH key: %v", err)
		res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultInvalidDNSyntax)
		w.Write(res)
		return
	}

	// Generate derived attributes
	attrs := derived.FromPublicKey(pubKey)

	keyType, fingerprint := getKeyInfo(pubKey)
	log.Printf("   Returning attributes for %s key %s (uid=%s, displayName=%s)",
		keyType, fingerprint, attrs.Username(), attrs.DisplayName("en"))

	// Return entry with derived attributes
	e := ldap.NewSearchResultEntry(dn)
	e.AddAttribute("objectClass", message.AttributeValue("inetOrgPerson"), message.AttributeValue("posixAccount"))
	e.AddAttribute("uid", message.AttributeValue(attrs.Username()))
	e.AddAttribute("uidNumber", message.AttributeValue(fmt.Sprintf("%d", attrs.PosixUserID())))
	e.AddAttribute("gidNumber", message.AttributeValue("1001")) // Constant group ID
	e.AddAttribute("homeDirectory", message.AttributeValue(fmt.Sprintf("/home/%s", attrs.Username())))
	e.AddAttribute("telephoneNumber", message.AttributeValue(attrs.PhoneNumber()))
	e.AddAttribute("displayName", message.AttributeValue(attrs.DisplayName("en")))
	e.AddAttribute("cn", message.AttributeValue(attrs.DisplayName("en"))) // Common Name

	// Generate locale-specific display names in this format:
	//	displayName;lang-zh: 用户123
	for _, lang := range attrs.SupportedLanguages() {
		e.AddAttribute(message.AttributeDescription(fmt.Sprintf("displayName;lang-%s", lang)), message.AttributeValue(attrs.DisplayName(lang)))
	}

	w.Write(e)

	log.Printf("✅ SEARCH COMPLETED: Returned 1 entry")
	res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
	w.Write(res)
}

func (s *LDAPServer) handleExtended(w ldap.ResponseWriter, m *ldap.Message) {
	extReq := m.GetExtendedRequest()
	clientAddr := m.Client.Addr().String()

	requestOID := string(extReq.RequestName())
	log.Printf("🔧 EXTENDED operation %s from %s", requestOID, clientAddr)

	// "Who Am I?" Extended Operation (RFC 4532)
	// OID: 1.3.6.1.4.1.4203.1.11.3
	const whoamiOID = "1.3.6.1.4.1.4203.1.11.3"

	if requestOID != whoamiOID {
		log.Printf("❌ EXTENDED REJECTED: Operation %s not supported", requestOID)
		res := ldap.NewExtendedResponse(ldap.LDAPResultUnwillingToPerform)
		res.SetDiagnosticMessage(fmt.Sprintf("Extended operation %s not supported", requestOID))
		w.Write(res)
		return
	}

	// Retrieve the bound DN from the session
	boundDNInterface, ok := s.sessions.Load(clientAddr)
	if !ok {
		log.Printf("❌ WHOAMI REJECTED: No bound session found for %s", clientAddr)
		res := ldap.NewExtendedResponse(ldap.LDAPResultOperationsError)
		res.SetDiagnosticMessage("Not authenticated - no bound session")
		w.Write(res)
		return
	}

	boundDN := boundDNInterface.(string)
	// RFC 4532: authzId format should be "dn:<distinguished-name>"
	authzId := "dn:" + boundDN

	log.Printf("✅ WHOAMI ACCEPTED: Returning authzId=%s", authzId)

	// Return the authorization identity in the response
	// Note: RFC 4532 specifies this should be in the response value, but the
	// goldap library may not support SetResponseValue. We'll use DiagnosticMessage
	// as a fallback that ldapwhoami can still display.
	res := ldap.NewExtendedResponse(ldap.LDAPResultSuccess)
	res.SetResponseName(whoamiOID)
	res.SetDiagnosticMessage(authzId)
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

// Start starts the LDAP server
func (s *LDAPServer) Start() error {
	return s.server.ListenAndServe(s.listenAddr)
}

// Stop stops the LDAP server
func (s *LDAPServer) Stop() {
	s.server.Stop()
}

// Addr returns the address the server is listening on
func (s *LDAPServer) Addr() string {
	if s.server.Listener != nil {
		return s.server.Listener.Addr().String()
	}
	return s.listenAddr
}
