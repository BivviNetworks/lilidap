package derived

import (
	"lilidap/internal/testutils"
	"regexp"
	"testing"
)

func TestUserAttributes(t *testing.T) {
	testKey, err := testutils.GetTestPublicKey()
	if err != nil {
		t.Fatal(err)
	}

	ua := FromPublicKey(testKey)

	t.Run("Username is consistent", func(t *testing.T) {
		username1 := ua.Username()
		username2 := ua.Username()
		if username1 != username2 {
			t.Errorf("Username not consistent: %s != %s", username1, username2)
		}
	})

	t.Run("Phone number is valid", func(t *testing.T) {
		phone := ua.PhoneNumber()
		// assert that the phone number is only digits using regex
		matched, _ := regexp.MatchString("^[0-9]+$", phone)
		if !matched {
			t.Errorf("Phone number not valid: %s", phone)
		}
		// assert that the phone number is the expected value based on the test key
		if phone != "82231210324213" {
			t.Errorf("Phone number not valid: %s", phone)
		}
	})
}
