package base32

import (
	"fmt"
	"math"
)

/*
 * This package encodes bits in a custom base32 character set for
 * safe and ergonomic use of hashed data
 */

// Omit confusable characters like in Base58, but also all the capitals and j and u
// Start with o and the digits to resemble hex, which aids debugging a bit
const base32Chars = "o123456789abcdefghikmnpqrstvwxyz"

// get a chunk of bits from a byte array, given a start bit and a chunk length n
func GetBits(input []byte, start int, n int) (uint, error) {
	if n > 8 || n < 1 {
		return 0, fmt.Errorf("chunk size must be 1<n<=8, got %d", n)
	}

	// Get bits from the start byte and possibly the following byte
	byteStart := start / 8
	bitStart := start % 8

	if byteStart >= len(input) {
		return 0, fmt.Errorf("start bit position in byte %d exceeds input length of %d", byteStart, len(input))
	}

	remainingBitsInFirstByte := 8 - bitStart
	mask := byte(0xFF >> bitStart)
	firstByteBits := uint(input[byteStart] & mask)
	if n <= remainingBitsInFirstByte { // All bits are in the first byte
		unneededBits := remainingBitsInFirstByte - n
		return firstByteBits >> unneededBits, nil
	}
	secondByteBitsNeeded := n - remainingBitsInFirstByte
	if byteStart+1 >= len(input) {
		return 0, fmt.Errorf("end bit position in byte %d exceeds input length of %d", byteStart+1, len(input))
	}
	secondByteBits := uint(input[byteStart+1] >> (8 - secondByteBitsNeeded))
	return (firstByteBits << secondByteBitsNeeded) | secondByteBits, nil
}

// encode a given length of bits to a base32 string
func Encode(input []byte, numBits int) (string, error) {
	outputLength := int(math.Ceil(float64(numBits) / 5))
	encoded := ""
	for i := 0; i < outputLength; i++ {
		value, err := GetBits(input, 5*i, 5)
		if err != nil {
			return "", err
		}
		encoded += string(base32Chars[value])
	}
	return encoded, nil

}
