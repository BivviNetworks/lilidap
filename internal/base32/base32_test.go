package base32_test

import (
	"lilidap/internal/base32"
	"testing"

	"github.com/stretchr/testify/require"
)

func compareBits(t *testing.T, input []byte, start int, n int, expected uint) {
	extracted, err := base32.GetBits(input, start, n)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, expected, extracted)
}

func compareEncode(t *testing.T, input []byte, numBits int, expected string) {
	encoded, err := base32.Encode(input, numBits)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, expected, encoded)
}

func TestBase32GetBits(t *testing.T) {
	compareBits(t, []byte{0xFF}, 0, 1, 0x01)
	compareBits(t, []byte{0xFF}, 0, 2, 0x03)
	compareBits(t, []byte{0xFF}, 0, 3, 0x07)
	compareBits(t, []byte{0xFF}, 0, 4, 0x0F)
	compareBits(t, []byte{0xFF}, 0, 8, 0xFF)
	compareBits(t, []byte{0xFF}, 0, 5, 0x1F)
	compareBits(t, []byte{0xFF}, 1, 5, 0x1F)
	compareBits(t, []byte{0xFF}, 1, 2, 0x03)
	compareBits(t, []byte{0xFF}, 2, 2, 0x03)
	compareBits(t, []byte{0xFF, 0xFF}, 6, 4, 0x0F)
	compareBits(t, []byte{0xFF, 0xFF}, 4, 5, 0x1F)
}

func TestBase32Encode(t *testing.T) {
	compareEncode(t, []byte{0xFF}, 5, "z")
	compareEncode(t, []byte{0xFF, 0xFF}, 5, "z")
	compareEncode(t, []byte{0xFF, 0xFF}, 10, "zz")
	compareEncode(t, []byte{0xA4, 0x36, 0x8B, 0xC4, 0x73}, 40, "mgv8qh3k")
}
