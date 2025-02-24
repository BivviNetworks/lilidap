# lilidap
LiliDAP is an LDAP proxy for hopping between networks, using SSH keys for authentication

## Motivation and Goals

The end goal is to support emergency communications on ad-hoc local area networks (WLANs) with no internet connection, providing services such as chat, VoIP calling, file sharing, and so on, but without the need for a centralized server administrator to perform the identity management that ordinarily underpins such services.  This would enable people (who likely have only their smartphones) to travel between local area networks without having to establish a distinct identity on each one, yet without relying on a centralized internet authority to prove who they are.

By adding this functionality to an identity management protocol (i.e. LDAP), we can enable chat, VoIP calling, and file sharing applications to delegate their login methods to this server.


## Mechanism

We will achieve identity management by using SSH keys (possession of a distinct private key) to establish identity instead of a password.  Usernames will be a public SSH key, and passwords will be the host and port of that user's SSH server (which proves possession of the private key).

## Integration

We currently envision this SSH-based management as an SASL mechanism, so that any application that supports SASL can use it.
