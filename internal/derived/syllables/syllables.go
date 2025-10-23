package syllables

import (
	"lilidap/internal/bitset"
	"math"
)

// SyllableGenerator defines rules for generating pronounceable syllables
type SyllableGenerator interface {
	// Indicate how many bits are representable in a single syllable, with or without coda
	BitsPerSyllable(fullSyllable bool) int
	// Generate creates a name using provided bits of entropy
	Generate(bits *bitset.BitSet) string
	// GenerateFromInt generates a syllable from an integer
	GenerateFromInt(i int) string
}

// BaseGenerator provides common functionality for syllable generation
type BaseGenerator struct {
	Onsets     []string // Initial consonant sounds
	Nuclei     []string // Vowel sounds (required)
	Codas      []string // Final consonant sounds (optional)
	onsetBits  int      // cached bit counts
	nucleiBits int      // cached bit counts
	codaBits   int      // cached bit counts
}

// BitsFor calculates bits needed to represent n choices
func BitsFor(choices []string) int {
	if len(choices) <= 1 {
		return 0
	}
	return int(math.Ceil(math.Log2(float64(len(choices)))))
}

// NewBaseGenerator creates a new BaseGenerator with the given onsets, nuclei, and codas
func NewBaseGenerator(onsets, nuclei, codas []string) BaseGenerator {
	if len(nuclei) == 0 {
		panic("nuclei are required")
	}

	return BaseGenerator{
		Onsets:     onsets,
		Nuclei:     nuclei,
		Codas:      codas,
		onsetBits:  BitsFor(onsets),
		nucleiBits: BitsFor(nuclei),
		codaBits:   BitsFor(codas),
	}
}

// Indicate how many bits are representable in a single syllable, with or without coda
func (g *BaseGenerator) BitsPerSyllable(fullSyllable bool) int {
	total := g.onsetBits + g.nucleiBits
	if fullSyllable {
		total += g.codaBits
	}
	return total
}

// Generate creates syllables from bits following phonetic rules
func (g *BaseGenerator) Generate(bits *bitset.BitSet) string {
	var result string
	bitPos := 0
	maxBits := bits.Size()

	for bitPos < maxBits {
		syllable := ""

		if g.onsetBits > 0 {
			onsetIdx := bits.Slice(bitPos, bitPos+g.onsetBits).ToInt()
			syllable += g.Onsets[onsetIdx]
			bitPos += g.onsetBits

		}

		// Always use nucleus
		nucleusIdx := bits.Slice(bitPos, bitPos+g.nucleiBits).ToInt()
		syllable += g.Nuclei[nucleusIdx]
		bitPos += g.nucleiBits

		// Optionally add coda if we have enough bits
		if g.codaBits > 0 {
			if bitPos < maxBits {
				codaIdx := bits.Slice(bitPos, bitPos+g.codaBits).ToInt()
				syllable += g.Codas[codaIdx]
				bitPos += g.codaBits
			}
		}

		result += syllable
	}

	return result
}

// EnglishGenerator implements English syllable rules
type EnglishGenerator struct {
	BaseGenerator
}

func NewEnglishGenerator() *EnglishGenerator {
	return &EnglishGenerator{
		NewBaseGenerator(
			[]string{"p", "t", "k", "b", "d", "g", "f", "v",
				"s", "z", "m", "n", "l", "r", "w", "y"},
			[]string{"a", "e", "i", "o", "u", "ai", "ei", "ou"},
			[]string{"m", "n", "l", "r", "s", "f", "t", "k"},
		),
	}
}
