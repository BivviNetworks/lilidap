module lilidap

go 1.21

require (
	github.com/go-ldap/ldap v0.0.0-00010101000000-000000000000
	github.com/go-ldap/ldap/v3 v3.4.10
	github.com/gopenguin/ldap-proxy v0.0.0-20171109213252-46570f88117c
	github.com/stretchr/testify v1.10.0
	github.com/thoas/go-funk v0.9.3
	github.com/vjeantet/ldapserver v1.0.1
	golang.org/x/crypto v0.31.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/lor00x/goldap v0.0.0-20180618054307-a546dffdd1a3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.8.1 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

// Replace old LDAP package with v3
replace github.com/go-ldap/ldap => github.com/go-ldap/ldap/v3 v3.4.5
