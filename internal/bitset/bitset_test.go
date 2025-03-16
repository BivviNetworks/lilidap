package bitset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitSetConstructionAndConversion(t *testing.T) {
	assert := assert.New(t)
	rawSize := 12
	a := New(rawSize)
	a.Set(0, false)
	a.Set(1, true)
	a.Set(2, false)
	a.Set(3, false)
	a.Set(4, true)
	a.Set(5, true)
	a.Set(6, false)
	a.Set(7, false)
	a.Set(8, false)
	a.Set(9, true)
	a.Set(10, true)
	a.Set(11, true)
	assert.Equalf(rawSize, a.Size(), "BitSet should have size %d, got %d", rawSize, a.Size())

	repBools := ([]bool{false, true, false, false, true, true, false, false, false, true, true, true})
	repBytes := []byte{0x0E, 0x32}
	repInt := 3634
	b := FromBools(repBools)
	c := FromBytes(repBytes, rawSize)
	d := FromInt(repInt, rawSize)

	assert.Truef(a.Equals(b), "Expected a(raw construction [%d] %d) == b(bool construction [%d] %d)",
		a.Size(), a.ToInt(), b.Size(), b.ToInt())

	assert.Truef(a.Equals(c), "Expected a(raw construction [%d] %d) == c(byte construction [%d] %d)",
		a.Size(), a.ToInt(), c.Size(), c.ToInt())

	assert.Truef(a.Equals(d), "Expected a(raw construction [%d] %d) == d(int construction [%d] %d)",
		a.Size(), a.ToInt(), d.Size(), d.ToInt())

	bools := a.ToBools()
	assert.Equalf(repBools, bools, "ToBools() should return the correct boolean array")

	bytes := a.ToBytes()
	assert.Equalf(repBytes, bytes, "ToBytes() should return the correct byte array")

	intVal := a.ToInt()
	assert.Equalf(repInt, intVal, "Expected a.ToInt() == %d, got %d", repInt, intVal)
}

func TestBitSetEqualAndConcat(t *testing.T) {
	assert := assert.New(t)
	a := FromBools([]bool{true, false, true})
	b := FromBools([]bool{true, false, true})
	c := FromBools([]bool{true, true, false})
	d := FromBools([]bool{true, false})

	// Test Equal
	assert.True(a.Equals(b), "Expected a == b")
	assert.False(a.Equals(c), "Expected a != c")
	assert.False(a.Equals(d), "Expected a != d (different sizes)")

	// Test Concat and Append
	ac := a.Append(c)
	assert.Equalf(6, ac.Size(), "Expected size 6, got %d", ac.Size())

	expected := []bool{true, false, true, true, true, false}
	assert.Equalf(expected, ac.ToBools(), "Concatenated BitSet should have correct bits")

	// Test multiple concat
	abc := Concat(a, b, c)
	assert.Equalf(9, abc.Size(), "Expected size 9, got %d", abc.Size())

	expected = []bool{true, false, true, true, false, true, true, true, false}
	assert.Equalf(expected, abc.ToBools(), "Multiple concatenated BitSet should have correct bits")

	// Test empty concat
	empty := New(0)
	aEmpty := a.Append(empty)
	assert.Truef(aEmpty.Equals(a), "Expected a + empty == a")

	emptyA := empty.Append(a)
	assert.Truef(emptyA.Equals(a), "Expected empty + a == a")

	// Test concat with no arguments
	justA := Concat(a)
	assert.Truef(justA.Equals(a), "Expected Concat(a) == a")
}

func TestBitSetBasic(t *testing.T) {
	assert := assert.New(t)
	bs := New(100)

	// Test Set and Get
	bs.SetBit(10)
	bs.SetBit(20)
	bs.SetBit(30)

	assert.True(bs.Get(10), "Expected bit 10 to be set")
	assert.True(bs.Get(20), "Expected bit 20 to be set")
	assert.True(bs.Get(30), "Expected bit 30 to be set")
	assert.False(bs.Get(15), "Expected bit 15 to be clear")

	// Test Clear
	bs.Clear(20)
	assert.False(bs.Get(20), "Expected bit 20 to be cleared")

	// Test Count
	assert.Equalf(2, bs.Count(), "Expected count 2, got %d", bs.Count())
}

func TestBitSetSlice(t *testing.T) {
	assert := assert.New(t)
	bs := FromBools([]bool{true, false, true, true, false, true, false, true})

	slice := bs.Slice(2, 6)
	assert.Equalf(4, slice.Size(), "Expected slice size 4, got %d", slice.Size())

	expected := []bool{true, true, false, true}
	assert.Equalf(expected, slice.ToBools(), "Slice should have correct bits")
}

func TestBitSetLogicalOps(t *testing.T) {
	assert := assert.New(t)
	a := FromBools([]bool{true, false, true, false})
	b := FromBools([]bool{false, true, true, false})

	// Test AND
	and := a.And(b)
	expected := []bool{false, false, true, false}
	assert.Equalf(expected, and.ToBools(), "AND operation should produce correct result")

	// Test OR
	or := a.Or(b)
	expected = []bool{true, true, true, false}
	assert.Equalf(expected, or.ToBools(), "OR operation should produce correct result")

	// Test XOR
	xor := a.Xor(b)
	expected = []bool{true, true, false, false}
	assert.Equalf(expected, xor.ToBools(), "XOR operation should produce correct result")
}

func TestStringConversions(t *testing.T) {
	assert := assert.New(t)
	a := FromBools([]bool{true, true, false, false, true, true, true, false, true})

	// test that the string is formatted correctly
	expectedRep := "BitSet(9)[0x0173]"
	assert.Equal(expectedRep, a.String())

	expectedStr := "101110011"
	assert.Equal(expectedStr, a.ToStringBinary())

	expectedOct := "563"
	assert.Equal(expectedOct, a.ToStringOctal())
}
