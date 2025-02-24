# lilidap
LiliDAP is an LDAP proxy for hopping between networks, using SSH keys for authentication

## Motivation and Goals

The end goal is to support emergency communications on ad-hoc local area networks (WLANs) with no internet connection, providing services such as chat, VoIP calling, file sharing, and so on, but without the need for a centralized server administrator to perform the identity management that ordinarily underpins such services.  This would enable people (who likely have only their smartphones) to travel between local area networks without having to establish a distinct identity on each one, yet without relying on a centralized internet authority to prove who they are.

By adding this functionality to an identity management protocol (i.e. LDAP), we can enable chat, VoIP calling, and file sharing applications to delegate their login methods to this server.


## Mechanism

We will achieve identity management by using SSH keys (possession of a distinct private key) to establish identity instead of a password.  Usernames will be a public SSH key, and passwords will be the host and port of that user's SSH server (which proves possession of the private key).


## Design

### LDAP DN Structure
The LDAP server uses a specific DN format to encode user identities:
- `cn=<base32-username>` : Contains the base32-encoded SSH public key
- `ou=campers` : Fixed string indicating temporary users
- `dc=0_1_0` : SemVer version number with underscores
- `dc=bivvi` : Fixed domain component

### Authentication Flow
1. Client sends BIND request:
   - DN contains SSH public key
   - Password contains host:port of SSH server
2. Server validates:
   - Decodes SSH key from DN
   - Verifies SSH server at host:port has matching key
3. On success, client can SEARCH to get derived attributes:
   - uid: Deterministic username from key
   - mobile: Phone number for VOIP
   - displayName: Human-readable name
   - objectClass: inetOrgPerson

### Client Integration
- VOIP clients use the mobile attribute
- Chat clients use uid and displayName
- File sharing uses uid for identity
