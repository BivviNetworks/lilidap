package derived

import (
	"github.com/stretchr/testify/assert"
	"lilidap/internal/testutils"
	"regexp"
	"testing"
)

func TestUserAttributes(t *testing.T) {
	assert := assert.New(t)
	testKey, err := testutils.GetTestPublicKey()
	if err != nil {
		t.Fatal(err)
	}

	ua := FromPublicKey(testKey)

	t.Run("Username is consistent", func(t *testing.T) {
		username1 := ua.Username()
		username2 := ua.Username()
		assert.Equalf(username1, username2, "Username not consistent: %s != %s", username1, username2)
	})

	t.Run("Username is properly generated", func(t *testing.T) {
		username := ua.Username()
		assert.Equal("uakmyrvel", username)
	})

	t.Run("PosixUserID is properly generated", func(t *testing.T) {
		uid := ua.PosixUserID()
		kb := ua.keyBits
		assert.Equal(kb.ToInt()+1000, uid, "PosixUserID should be the int val of the hash plus 1000")
	})

	t.Run("Phone number is valid", func(t *testing.T) {
		phone := ua.PhoneNumber()
		// assert that the phone number is only digits using regex
		matched, _ := regexp.MatchString("^8[0-9]+$", phone)
		assert.Truef(matched, "Phone number starting with 8 not valid: %s", phone)

		// assert that the phone number is the expected value based on the test key
		assert.Equal("8930577239884", phone)
	})

	t.Run("Generated name is valid", func(t *testing.T) {
		name := ua.DisplayName("en")
		assert.Equal("lutbousnifkeit", name)
	})

	t.Run("Supported languages", func(t *testing.T) {
		langs := ua.SupportedLanguages()
		assert.Equal([]string{"en"}, langs)
	})
}
