# lilidap
LiliDAP is an LDAP proxy for hopping between networks, using SSH keys for authentication

## Motivation and Goals

The end goal is to support emergency communications on ad-hoc local area networks (WLANs) with no internet connection, providing services such as chat, VoIP calling, file sharing, and so on, but without the need for a centralized server administrator to perform the identity management that ordinarily underpins such services.  This would enable people (who likely have only their smartphones) to travel between local area networks without having to establish a distinct identity on each one, yet without relying on a centralized internet authority to prove who they are.

By adding this functionality to an identity management protocol (i.e. LDAP), we can enable chat, VoIP calling, and file sharing applications to delegate their login methods to this server.


## Mechanism

We will achieve identity management by using SSH keys (possession of a distinct private key) to establish identity instead of a password. Users authenticate by providing their **full SSH public key** as the username and the **host:port of their SSH server** as the password. The LDAP server validates key ownership by connecting to that SSH server.

## Running the Server

### Usage Examples

**Production (default):**
```bash
sudo ./lilidap
# Listens on: 0.0.0.0:389 (all interfaces, standard LDAP port)
# Note: Port 389 requires root/administrator privileges
```

**Testing mode (no root needed):**
```bash
./lilidap --host localhost --port 3389
# Listens on: localhost:3389 (safe for testing)
```

**Custom port on all interfaces:**
```bash
./lilidap --port 10389
# Listens on: 0.0.0.0:10389 (all interfaces, custom port)
```

**Specific network interface:**
```bash
./lilidap --host 192.168.1.100
# Listens on: 192.168.1.100:389 (specific IP, default port)
```

**Full customization:**
```bash
./lilidap --host 10.0.0.5 --port 8389
# Listens on: 10.0.0.5:8389 (specific IP and port)
```

### Testing the Server

Once the server is running, test it with `ldapwhoami`:

```bash
# Get your SSH public key
MY_KEY=$(cat ~/.ssh/id_rsa.pub)

# Test authentication (adjust host:port as needed)
ldapwhoami -H ldap://localhost:3389 \
  -D "cn=${MY_KEY},ou=campers,dc=0_1_0,dc=bivvi" \
  -w "127.0.0.1:22"
```

See [TESTING.md](TESTING.md) for comprehensive testing instructions.

## Design

### Three Types of Identifiers

LiliDAP uses three different identifiers for each user, all derived from their SSH public key:

1. **Login Credentials** (for authentication)
   - Username: Full SSH public key in OpenSSH format (e.g., `ssh-rsa AAAAB3NzaC1yc2E...`)
   - Password: `host:port` where their SSH server runs (e.g., `192.168.1.100:22`)
   - Purpose: Proves identity through SSH key ownership

2. **POSIX Username** (for system identification)
   - Format: `u` + 8-char base32-encoded hash (e.g., `u1234abcd`)
   - Character set: `o123456789abcdefghikmnpqrstvwxyz` (omits confusables)
   - Purpose: File ownership, process IDs, system-level identification
   - Stored in: `uid` attribute

3. **Friendly Name** (for human display)
   - Format: Syllabic encoding (e.g., `vantumkeirrof`)
   - Purpose: Display in chat, caller ID, user directories
   - Stored in: `displayName` attribute

### LDAP DN Structure

The LDAP server uses a specific DN format with the full SSH public key:
- `cn=<ssh-public-key>` : Full SSH public key in OpenSSH authorized_keys format
  - Example: `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...`
- `ou=campers` : Fixed string indicating temporary/mobile users
- `dc=0_1_0` : SemVer version number with underscores
- `dc=bivvi` : Fixed domain component

Complete DN example:
```
cn=ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...,ou=campers,dc=0_1_0,dc=bivvi
```

### Authentication Flow

#### For Direct LDAP Access:
1. Client sends BIND request:
   - DN: Contains full SSH public key as CN
   - Password: `host:port` of their SSH server
2. Server validates:
   - Extracts SSH public key from DN's CN
   - Connects to the SSH server at `host:port`
   - Verifies that server presents the same public key
3. On success: Authentication succeeds, record created on-demand

#### For Service Integration (FreePBX, IRC, etc.):
1. User provides to service:
   - Username: Full SSH public key
   - Password: Their SSH server's `host:port`
2. Service performs LDAP BIND using those credentials
3. On success, service performs LDAP SEARCH to get user attributes:
   - `uid`: POSIX username (e.g., `u1234abcd`)
   - `displayName`: Friendly name (e.g., `vantumkeirrof`)
   - `telephoneNumber`: Derived phone number for VoIP
   - `uidNumber`: Numeric user ID
   - `gidNumber`: Group ID (constant: 1001)
   - `homeDirectory`: Home directory path
4. Service uses these attributes:
   - VoIP shows `vantumkeirrof` as caller ID
   - Chat displays the friendly name
   - File server uses `u1234abcd` for ownership

### On-Demand Record Creation

Records are created dynamically when a user successfully authenticates:
- No pre-registration required
- All attributes derived deterministically from SSH key hash
- Same SSH key always produces the same identity
- Enables "network hopping" without central authority

### Identity Consistency ("Hopping")

When a user moves between networks:
1. They keep their SSH private key
2. They spin up an SSH server on the new network
3. They authenticate with: SSH key + new host:port
4. LDAP derives the **same** attributes:
   - Same `uid` (e.g., `u1234abcd`)
   - Same `displayName` (e.g., `vantumkeirrof`)
   - Same `telephoneNumber`
5. Result: Identity is preserved across networks without coordination

### Derived Attributes

All user attributes are deterministically derived from the SHA-256 hash of the SSH public key:

- `objectClass`: `inetOrgPerson`, `posixAccount`
- `uid`: Base32-encoded hash prefix (POSIX username) - e.g., `u1234abcd`
- `uidNumber`: Integer derived from hash (starting at 1000)
- `gidNumber`: Constant value (1001)
- `homeDirectory`: Constructed from uid (e.g., `/home/u1234abcd`)
- `telephoneNumber`: Formatted hash value for VoIP routing
- `displayName`: Syllable-generated pronounceable name
- `displayName;lang-XX`: Locale-specific variants
- `cn`: Copy of display name

#### Base32 Encoding Details

The base32 encoding used for `uid` generation:
- Character set: `o123456789abcdefghikmnpqrstvwxyz`
- Omits confusable characters (like `0`/`O`, `1`/`l`/`I`)
- Applied only to the UID attribute, not the DN
- Example: SSH key hash → `u1234abcd` (8 characters after 'u' prefix)

### Client Integration
- **VoIP clients** (FreePBX): Use `telephoneNumber` for call routing, `displayName` for caller ID
- **Chat clients** (IRC): Use `uid` for unique identity, `displayName` for display
- **File sharing**: Use `uid` and `uidNumber` for POSIX permissions
- **All services**: Authenticate users via LDAP BIND with SSH credentials
