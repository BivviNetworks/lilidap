package derived

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"math"
	"strings"

	"golang.org/x/crypto/ssh"
)

const KEY_HASH_BIT_LENGTH = 40

// UserAttributes generates consistent LDAP attributes from an SSH public key
type UserAttributes struct {
	pubKey ssh.PublicKey
	hash   [32]byte // cached sha256 of pubKey
}

// FromPublicKey creates an attribute generator from an SSH public key
func FromPublicKey(pubKey ssh.PublicKey) *UserAttributes {
	return &UserAttributes{
		pubKey: pubKey,
		hash:   sha256.Sum256(pubKey.Marshal()),
	}
}

// Username returns a consistent unique identifier
func (ua *UserAttributes) Username() string {
	b32str := strings.ToLower(base32.StdEncoding.EncodeToString(ua.hash[:])[:8])
	return fmt.Sprintf("u%s", b32str)
}

// PhoneNumber returns a consistent unique phone number
func (ua *UserAttributes) PhoneNumber() string {
	digits := ua.encodeOctal(KEY_HASH_BIT_LENGTH)
	return fmt.Sprintf("8%s", digits)
}

// DisplayName returns a locale-specific display name
func (ua *UserAttributes) DisplayName(locale string) string {
	// TODO: Implement locale-specific name generation
	return ua.Username()
}

// encodeOctal converts the first numBits of hash to octal
func (ua *UserAttributes) encodeOctal(numBits int) string {
	outputLength := int(math.Ceil(float64(numBits) / 8))
	encoded := ""
	for _, b := range ua.hash[0:outputLength] {
		encoded += fmt.Sprintf("%o", b)
	}
	return encoded
}
