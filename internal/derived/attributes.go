package derived

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"

	"lilidap/internal/bitset"
	"lilidap/internal/derived/syllables"

	"golang.org/x/crypto/ssh"
)

// There are just too many bits in the fingerprint to use all of them
// for derived attributes, so we will attempt to pick a number that will
// allow roughly 1 million people to use this system without collision.
//
// For 40 bits, this is calculated as the square root of 2^40 = 1,048,576
const KEY_HASH_BIT_LENGTH = 40

// UserAttributes generates consistent LDAP attributes from an SSH public key
type UserAttributes struct {
	pubKey       ssh.PublicKey
	hash         [32]byte       // cached sha256 of pubKey
	keyBits      *bitset.BitSet // KEY_HASH_BIT_LENGTH bits from the hash
	displayNames map[string]string
}

// FromPublicKey creates an attribute generator from an SSH public key
func FromPublicKey(pubKey ssh.PublicKey) *UserAttributes {
	// Get deterministic seed from key fingerprint
	theHash := sha256.Sum256(pubKey.Marshal())
	attrs := &UserAttributes{
		pubKey:       pubKey,
		hash:         theHash,
		keyBits:      bitset.FromBytes(theHash[:], KEY_HASH_BIT_LENGTH),
		displayNames: make(map[string]string),
	}

	// Generate English name using syllables

	attrs.displayNames["en"] = syllables.NewEnglishGenerator().Generate(attrs.keyBits)

	return attrs
}

// Username returns a consistent unique identifier
func (ua *UserAttributes) Username() string {
	b32str := strings.ToLower(base32.StdEncoding.EncodeToString(ua.hash[:])[:8])
	return fmt.Sprintf("u%s", b32str)
}

// UserID starting at 1000
func (ua *UserAttributes) PosixUserID() int {
	return ua.keyBits.ToInt() + 1000
}

// PhoneNumber returns a consistent unique phone number
func (ua *UserAttributes) PhoneNumber() string {
	return fmt.Sprintf("8%0*d", 4, ua.keyBits.ToInt())
}

// DisplayName returns a locale-specific display name
func (ua *UserAttributes) DisplayName(lang string) string {
	if name, ok := ua.displayNames[lang]; ok {
		return name
	}
	// Fall back to English if language not supported
	return ua.displayNames["en"]
}

// SupportedLanguages returns a list of supported languages
//
// This drives LDAP support like the following:
//
//	displayName: User-123        # Default/fallback
//	displayName;lang-en: User-123
//	displayName;lang-ja: ユーザー123
//	displayName;lang-zh: 用户123
func (ua *UserAttributes) SupportedLanguages() []string {
	langs := make([]string, 0, len(ua.displayNames))
	for lang := range ua.displayNames {
		langs = append(langs, lang)
	}
	return langs
}
