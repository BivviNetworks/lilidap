# LiliDAP Identity Manager

## Purpose

A standalone application that provides a complete, zero-configuration identity solution:

1. Manages a persistent Ed25519 SSH key pair (or loads an existing one)
2. Starts an SSH server on a free port
3. Displays your LDAP credentials (DN + password) for copy/paste into any LDAP-enabled service
4. Shows your derived identity (username, display name, phone number)

## Usage

```bash
# Start identity manager (uses/creates ~/.lilidap/identity)
lilidap-identity

# Use a different SSH key location
lilidap-identity --key ~/my-identity

# Use your existing SSH key
lilidap-identity --key ~/.ssh/id_ed25519

# Specify SSH port (instead of random)
lilidap-identity --ssh-port 2222

# Bind to specific interface
lilidap-identity --ssh-host 192.168.1.100
```

## Output Example

```
╔═══════════════════════════════════════════════════════════════╗
║           LiliDAP Identity Manager                            ║
╚═══════════════════════════════════════════════════════════════╝

🔑 Identity Ready!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 COPY THESE CREDENTIALS:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Username (DN):
cn=ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...,ou=campers,dc=0_1_0,dc=bivvi

Password:
127.0.0.1:34567

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
👤 YOUR DERIVED IDENTITY:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

POSIX Username:  u1m4k7p2      (for file systems, IRC, etc.)
Friendly Name:   vantumkeirrof (for display, caller ID)
Phone Number:    8753198499    (for VoIP)
User ID:         753198499     (numeric UID)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🌐 SSH Server: Running on 127.0.0.1:34567
🔐 SSH Key: /tmp/lilidap-identity-abc123/id_rsa

✅ Ready to authenticate! Use these credentials with any LDAP-enabled service.

Press Ctrl+C to stop.
```

## Detailed Specification

### Command-Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--key` | string | `~/.lilidap/identity` | Path to SSH private key. If file doesn't exist, generates new Ed25519 key at this location. |
| `--ssh-host` | string | `127.0.0.1` | Host/IP to bind SSH server to |
| `--ssh-port` | int | auto | SSH server port. If 0 or not specified, finds a free port automatically. |

### Key Management Strategy

The identity manager uses a persistent Ed25519 key by default, automatically creating it on first run.

#### Scenario 1: Default Key (First Run)
```bash
lilidap-identity
# Generates ~/.lilidap/identity (Ed25519) if it doesn't exist
# Creates both identity (private, 0600) and identity.pub (public, 0644)
# Good for: Normal usage, persistent identity across network hops
```

#### Scenario 2: Default Key (Subsequent Runs)
```bash
lilidap-identity
# Loads existing ~/.lilidap/identity
# Same identity every time - consistent username, display name, etc.
# Good for: Daily use, network hopping with same identity
```

#### Scenario 3: Custom Key Location
```bash
lilidap-identity --key ~/my-special-identity
# First run: Generates Ed25519 key at ~/my-special-identity
# Later runs: Loads the existing key
# Good for: Multiple identities, testing different personas
```

#### Scenario 4: Use Existing SSH Key
```bash
lilidap-identity --key ~/.ssh/id_ed25519
# Loads your existing SSH key
# Works with Ed25519, RSA, ECDSA
# Good for: Using established SSH identity, no new key needed
```

#### Scenario 5: Use System Host Key
```bash
sudo lilidap-identity --key /etc/ssh/ssh_host_ed25519_key
# Loads system's SSH host key
# Good for: Server installations, stable machine identity
```

**Key Type**: New keys are always Ed25519 (modern, secure, fast). Existing keys of any type (RSA, ECDSA, Ed25519) can be loaded.

### Port Selection Strategy

Based on `tcp_helpers.GetFreePort()` and `WithSSHServer()` pattern:

1. **Automatic (default)**: Use `net.Listen("tcp", "127.0.0.1:0")` to get OS-assigned port
2. **Manual**: User specifies `--ssh-port 2222`, bind to that specific port
3. **Port validation**: Check if manually specified port is available before starting
4. **Error handling**: If specified port unavailable, exit with clear error (don't auto-fallback)

### SSH Server Behavior

The SSH server should:

1. **Accept connections** but not allow successful authentication
   - Similar to `SampleServerConfigs["AuthPassword"]` pattern
   - Provides password callback that always rejects
   - This allows key validation via connection attempt

2. **Host key**: Use the identity key as the server's host key
   - Clients connecting will see this key during handshake
   - This is what LiliDAP validates for identity

3. **No actual shell access**: Reject all channels
   - Following the pattern in `StartMockSSHServer`
   - Server only needs to present its host key, not provide access

4. **Minimal auth methods**:
   ```go
   config := &ssh.ServerConfig{
       MaxAuthTries: 1,
       PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
           return nil, fmt.Errorf("password authentication not supported")
       },
   }
   config.AddHostKey(signer)
   ```

### Public Key Normalization

To ensure consistent representation (per ldapserver.go changes):

```go
// After loading/generating key
pubKey := signer.PublicKey()

// Normalize to canonical form (no trailing whitespace/newlines)
normalizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubKey)))

// Use normalized key for DN construction
dn := fmt.Sprintf("cn=%s,ou=campers,dc=0_1_0,dc=bivvi", normalizedKey)
```

### Error Handling

The application should handle:

- **Key file not found**: Only when explicitly specified with --key (default creates new)
- **Invalid key format**: Indicate what's wrong with the key file
- **Port already in use**: Show port number and suggest alternatives
- **Permission denied**: Especially for system keys, low ports, or ~/.lilidap directory creation
- **Directory creation failure**: Can't create ~/.lilidap
- **Key save failure**: Can't write key files
- **Unsupported key type**: When loading (though all standard SSH keys work)

### Implementation Plan

#### Phase 1: Basic Identity Manager (Current Priority)

```go
// cmd/lilidap-identity/main.go
func main() {
    // 1. Parse flags
    keyPath := flag.String("key", "~/.lilidap/identity", "Path to SSH private key")
    sshHost := flag.String("ssh-host", "127.0.0.1", "SSH server host")
    sshPort := flag.Int("ssh-port", 0, "SSH server port (0=auto)")
    flag.Parse()

    // 2. Expand home directory in key path
    expandedKeyPath, err := expandPath(*keyPath)
    if err != nil {
        log.Fatalf("Invalid key path: %v", err)
    }

    // 3. Load or generate SSH key pair (auto-saves if generated)
    signer, pubKey, err := getOrCreateKey(expandedKeyPath)
    if err != nil {
        log.Fatalf("Key management error: %v", err)
    }

    // 4. Find a free port (or use specified)
    port, err := getPort(*sshHost, *sshPort)
    if err != nil {
        log.Fatalf("Port selection error: %v", err)
    }

    // 5. Start SSH server
    stopServer, err := startSSHServer(signer, *sshHost, port)
    if err != nil {
        log.Fatalf("SSH server error: %v", err)
    }
    defer stopServer()

    // 6. Derive identity attributes
    attrs := derived.FromPublicKey(pubKey)

    // 7. Display credentials
    displayCredentials(pubKey, *sshHost, port, attrs, expandedKeyPath)

    // 8. Keep running until interrupted
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    fmt.Println("\n\n👋 Shutting down gracefully...")
}
```

### Phase 2: GUI (Future)

- Desktop app with copy buttons
- QR code for mobile scanning
- System tray icon
- Clipboard integration

## Files Structure

```
cmd/lilidap-identity/
├── README.md (this file)
├── main.go               # Entry point, flag parsing, orchestration
├── ssh_server.go         # SSH server implementation
├── key_manager.go        # SSH key generation/loading/saving
├── display.go            # Terminal UI and credential display
└── port_manager.go       # Port selection and validation
```

### Module Responsibilities

#### `main.go`
- Parse command-line flags
- Coordinate all components
- Handle signals (Ctrl+C)
- Exit with appropriate codes

#### `key_manager.go`
```go
// getOrCreateKey loads existing key or generates new Ed25519 key at keyPath
// Always saves newly generated keys to disk
// Returns signer and public key
func getOrCreateKey(keyPath string) (ssh.Signer, ssh.PublicKey, error)

// loadPrivateKey loads an SSH private key from disk
// Supports Ed25519, RSA, ECDSA formats
func loadPrivateKey(path string) (ssh.Signer, error)

// generateEd25519Key creates a new Ed25519 key pair
func generateEd25519Key() (ssh.Signer, error)

// saveKey writes private key to disk (OpenSSH format, 0600 permissions)
// Also writes public key to path.pub (0644 permissions)
func saveKey(signer ssh.Signer, path string) error

// expandPath expands ~ to home directory
func expandPath(path string) (string, error)
```

#### `ssh_server.go`
```go
// startSSHServer starts an SSH server and returns a stop function
// Based on StartMockSSHServer from internal/testutils/ssh_helpers
func startSSHServer(signer ssh.Signer, host string, port int) (stopFunc func(), err error)

// Creates minimal config that:
// - Uses signer as host key
// - Rejects all authentication attempts
// - Rejects all channel requests
// - Logs connections (optional debug mode)
```

#### `port_manager.go`
```go
// getPort returns an available port
// If requestedPort is 0, finds a free port automatically
// If requestedPort is specified, validates it's available
func getPort(host string, requestedPort int) (int, error)

// Based on tcp_helpers.GetFreePort but with validation
```

#### `display.go`
```go
// displayCredentials shows the formatted output
func displayCredentials(pubKey ssh.PublicKey, host string, port int, attrs *derived.Attributes, keyPath string)

// Displays:
// - Banner
// - LDAP DN and password
// - Derived identity (username, display name, phone, UID)
// - SSH server info
// - Key file location (helpful for first run)
// - Instructions
```

## Dependencies

- `golang.org/x/crypto/ssh` - SSH server and key management
- `lilidap/internal/derived` - Identity derivation
- `lilidap/internal/testutils/tcp_helpers` - Port selection utilities

## Use Cases

### 1. Quick Testing
```bash
# Terminal 1: Start LDAP server
cd cmd/lilidap && go run main.go --host localhost --port 3389

# Terminal 2: Start identity manager
cd cmd/lilidap-identity && go run *.go

# Terminal 3: Test LDAP auth (copy credentials from Terminal 2)
ldapwhoami -H ldap://localhost:3389 -D "<DN>" -w "127.0.0.1:<port>"
```

### 2. Production Use
```bash
# User runs on their laptop
lilidap-identity

# Copy credentials to FreePBX/IRC/etc.
# Services authenticate against organization's LDAP server
# Identity persists at ~/.lilidap/identity
```

### 3. Network Hopping
```bash
# Network A: Start identity manager
lilidap-identity

# Move to Network B: Start with same key (auto-loads from ~/.lilidap/identity)
lilidap-identity

# Same username/friendly name/phone everywhere!
# Your identity follows you across networks
```

### 4. Multiple Identities
```bash
# Work identity
lilidap-identity --key ~/.lilidap/work-identity

# Personal identity
lilidap-identity --key ~/.lilidap/personal-identity

# Each has different username, display name, phone number
```

## Testing & Validation

### Manual Testing Scenarios

#### Test 1: Auto-generate temporary identity
```bash
./lilidap-identity
# Verify: Shows DN, password, derived attributes
# Verify: SSH server accepts connections on displayed port
# Verify: ldapwhoami succeeds (if LDAP server available)
# Verify: Ctrl+C shuts down gracefully
```

#### Test 2: First run with default key
```bash
rm -rf ~/.lilidap  # Clean slate
./lilidap-identity
# Verify: Creates ~/.lilidap/identity and ~/.lilidap/identity.pub
# Verify: Files have correct permissions (0600 for private, 0644 for public)
# Verify: Shows "Generated new Ed25519 key" message

# Run again
./lilidap-identity
# Verify: Loads existing key (same DN, same derived attributes)
# Verify: Shows "Using existing key" message
```

#### Test 3: Custom key location
```bash
./lilidap-identity --key /tmp/custom-identity
# First run - verify: Generates new key at /tmp/custom-identity
# Verify: Creates /tmp/custom-identity and /tmp/custom-identity.pub

./lilidap-identity --key /tmp/custom-identity
# Second run - verify: Loads existing key
# Verify: Same identity as first run
```

#### Test 4: Use existing SSH key
```bash
# Use your real SSH key
./lilidap-identity --key ~/.ssh/id_ed25519
# Verify: Loads the key successfully
# Verify: DN matches the public key
# Verify: Derived attributes consistent
```

#### Test 5: Manual port selection
```bash
./lilidap-identity --ssh-port 2222
# Verify: SSH server runs on port 2222
# Verify: Password shows 127.0.0.1:2222

# Try with port in use:
./lilidap-identity --ssh-port 22  # (likely in use)
# Verify: Clear error about port unavailable
```

#### Test 6: Different hosts
```bash
./lilidap-identity --ssh-host 0.0.0.0 --ssh-port 2222
# Verify: SSH server accessible from other machines
# Verify: Password reflects actual network interface
```

### Automated Tests

Create `cmd/lilidap-identity/*_test.go`:

```go
// key_manager_test.go
func TestGenerateEd25519Key(t *testing.T) { ... }
func TestLoadExistingKey(t *testing.T) { ... }
func TestSaveAndLoadKey(t *testing.T) { ... }
func TestExpandPath(t *testing.T) { ... }

// port_manager_test.go
func TestGetFreePort(t *testing.T) { ... }
func TestGetSpecificPort(t *testing.T) { ... }
func TestPortInUse(t *testing.T) { ... }

// ssh_server_test.go
func TestSSHServerStartStop(t *testing.T) { ... }
func TestSSHServerPresentsKey(t *testing.T) { ... }
```

### Integration Test with LDAP Server

```bash
# Terminal 1: Start identity manager
./lilidap-identity

# Terminal 2: Start LDAP server
../lilidap/lilidap --host localhost --port 3389

# Terminal 3: Copy credentials from Terminal 1 and test
ldapwhoami -H ldap://localhost:3389 -D "<DN>" -w "<host:port>"
# Expected: Success with normalized DN returned

ldapsearch -H ldap://localhost:3389 -D "<DN>" -w "<host:port>" \
  -b "<DN>" "(objectClass=*)"
# Expected: Returns user attributes (uid, displayName, telephoneNumber, etc.)
```

### Success Criteria

Before considering Phase 1 complete:

- ✅ Generates Ed25519 keys by default
- ✅ Loads existing keys (Ed25519, RSA, ECDSA)
- ✅ Saves keys with correct permissions (0600 private, 0644 public)
- ✅ Expands ~ in paths correctly
- ✅ Creates ~/.lilidap directory if needed
- ✅ Finds free ports automatically
- ✅ Validates manually specified ports
- ✅ SSH server accepts connections
- ✅ SSH server presents host key correctly
- ✅ Displays normalized public key in DN
- ✅ Shows derived attributes matching `internal/derived`
- ✅ Shows key file location in output
- ✅ Graceful shutdown on Ctrl+C
- ✅ Clear error messages for common issues
- ✅ All automated tests pass
- ✅ All manual test scenarios work

## Next Steps

1. ✅ Fix main.go for basic LDAP server
2. ✅ Create detailed specification (this document)
3. ⏳ Implement `key_manager.go` (Ed25519 generation, loading, saving)
4. ⏳ Implement `port_manager.go` (auto/manual port selection)
5. ⏳ Implement `ssh_server.go` (minimal SSH server)
6. ⏳ Implement `display.go` (credential display)
7. ⏳ Implement `main.go` (Phase 1)
8. ⏳ Add automated tests
9. ⏳ Validate against success criteria
10. ⏳ Update CI/CD to build lilidap-identity
