package syllables

import (
	"lilidap/internal/bitset"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseGenerator(t *testing.T) {
	assert := assert.New(t)

	t.Run("Bits calculation", func(t *testing.T) {
		gen := NewBaseGenerator(
			make([]string, 16), // 4 bits
			make([]string, 8),  // 3 bits
			make([]string, 8),  // 3 bits
		)
		// expect 7 bits for partial syllable and 10 bits for full syllable
		assert.Equalf(7, gen.BitsPerSyllable(false), "Expected 7 bits for partial syllable, got %d", gen.BitsPerSyllable(false))
		assert.Equalf(10, gen.BitsPerSyllable(true), "Expected 10 bits for full syllable, got %d", gen.BitsPerSyllable(true))
	})

	t.Run("Empty lists", func(t *testing.T) {
		gen := NewBaseGenerator(
			make([]string, 0), []string{"a"}, make([]string, 0), // Single nucleus = 0 bits
		)
		assert.Equalf(0, gen.BitsPerSyllable(false), "Expected 0 bits for partial syllable, got %d", gen.BitsPerSyllable(false))
		assert.Equalf(0, gen.BitsPerSyllable(true), "Expected 0 bits for full syllable, got %d", gen.BitsPerSyllable(true))
	})

	t.Run("Nucleus required", func(t *testing.T) {
		gen := BaseGenerator{}
		defer func() {
			if r := recover(); r == nil {
				assert.Fail("Expected panic with empty nuclei")
			}
		}()
		gen.Generate(bitset.FromBytes([]byte{0}))
	})
}

func TestEnglishGenerator(t *testing.T) {
	assert := assert.New(t)
	gen := NewEnglishGenerator()

	t.Run("Syllable structure", func(t *testing.T) {

		// check the bits per syllable for partial and full syllables
		// then for each of those lists, iterate over the entire range of possible syllables
		// and validate that the syllable is valid

		for _, fullSyllable := range []bool{false, true} {
			bits := gen.BitsPerSyllable(fullSyllable)
			for i := 0; i < 1<<bits; i++ {
				// directly generate the syllable
				syllable := gen.Generate(bitset.FromInt(i, bits))
				assert.Truef(isValidEnglishSyllable(syllable, gen), "Invalid syllable: %s", syllable)
			}
		}
	})

	t.Run("Deterministic", func(t *testing.T) {
		bs := bitset.FromBytes([]byte{0xA5, 0x5A})
		name1 := gen.Generate(bs)
		name2 := gen.Generate(bs)
		assert.Equalf(name1, name2, "Not deterministic: %s != %s", name1, name2)
	})
}

func isValidEnglishSyllable(syllable string, gen *EnglishGenerator) bool {
	// Check if syllable follows onset-nucleus-coda pattern
	for _, onset := range gen.Onsets {
		if strings.HasPrefix(syllable, onset) {
			rest := syllable[len(onset):]
			for _, nucleus := range gen.Nuclei {
				if strings.HasPrefix(rest, nucleus) {
					final := rest[len(nucleus):]
					if len(final) == 0 {
						return true
					}
					for _, coda := range gen.Codas {
						if final == coda {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// check some generated 40 bit names, we will hand-pick a few random values
func TestSampleGenerations(t *testing.T) {
	assert := assert.New(t)
	gen := NewEnglishGenerator()

	// Define test cases as a map of input integers to expected output strings
	testCases := map[int64]string{
		0xDEADBEEF12: "ketnoknaifmout",
		0x123456890:  "penmaildosdam",
		0:            "pampampampam",
		0xFFFFFFFFFF: "youkyoukyoukyouk",
		0xCAFEBABE:   "wofwikyinbam",
	}

	// Iterate through test cases
	for input, expected := range testCases {
		result := gen.Generate(bitset.FromInt(int(input), 40))
		assert.Equalf(expected, result, "For input 0x%X, expected %s, got %s", input, expected, result)
	}
}
