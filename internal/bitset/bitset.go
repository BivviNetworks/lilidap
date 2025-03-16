package bitset

import (
	"fmt"
	"math/bits"
)

// BitSet represents an arbitrary-length array of bits
type BitSet struct {
	bits []uint64
	size int
}

// New creates a new BitSet with the specified size
func New(size int) *BitSet {
	return &BitSet{
		bits: make([]uint64, (size+63)/64),
		size: size,
	}
}

// FromBools creates a BitSet from a boolean array
func FromBools(bools []bool) *BitSet {
	bs := New(len(bools))
	for i, val := range bools {
		bs.Set(i, val)
	}
	return bs
}

// FromBytes creates a BitSet from a byte array with optional length
func FromBytes(bytes []byte, length ...int) *BitSet {
	size := len(bytes) * 8
	if len(length) > 0 && length[0] > 0 && length[0] <= size {
		size = length[0]
	}

	bs := New(size)
	for i := 0; i < size; i++ {
		bytePos, bitPos := len(bytes)-1-i/8, i%8
		bs.Set(i, bytes[bytePos]&(1<<bitPos) != 0)
	}
	return bs
}

func FromInt(num int, size int) *BitSet {
	bs := New(size)
	for i := 0; i < size; i++ {
		bs.Set(i, num&(1<<i) != 0)
	}
	return bs
}

// Size returns the number of bits in the BitSet
func (bs *BitSet) Size() int {
	return bs.size
}

// Set sets the bit at position i to the given value
func (bs *BitSet) Set(i int, val bool) {
	if i >= bs.size || i < 0 {
		panic(fmt.Sprintf("BitSet index out of range: %d (size: %d)", i, bs.size))
	}
	idx, bit := i/64, i%64
	if val {
		bs.bits[idx] |= 1 << bit
	} else {
		bs.bits[idx] &= ^(1 << bit)
	}
}

// SetBit sets the bit at position i to 1 (for backward compatibility)
func (bs *BitSet) SetBit(i int) {
	bs.Set(i, true)
}

// Clear sets the bit at position i to 0
func (bs *BitSet) Clear(i int) {
	bs.Set(i, false)
}

// Get returns the bit value at position i
func (bs *BitSet) Get(i int) bool {
	if i >= bs.size || i < 0 {
		panic(fmt.Sprintf("BitSet index out of range: %d (size: %d)", i, bs.size))
	}
	idx, bit := i/64, i%64
	return (bs.bits[idx] & (1 << bit)) != 0
}

// get the bit as an int
func (bs *BitSet) GetInt(i int) int {
	if i >= bs.size || i < 0 {
		panic(fmt.Sprintf("BitSet index out of range: %d (size: %d)", i, bs.size))
	}
	if bs.Get(i) {
		return 1
	}
	return 0
}

// Slice returns a new BitSet containing bits from start to end (exclusive)
func (bs *BitSet) Slice(start, end int) *BitSet {
	if start < 0 {
		start = 0
	}
	if end > bs.size {
		end = bs.size
	}
	if start >= end {
		return New(0)
	}

	result := New(end - start)
	for i := start; i < end; i++ {
		result.Set(i-start, bs.Get(i))
	}
	return result
}

// And performs a bitwise AND with another BitSet
func (bs *BitSet) And(other *BitSet) *BitSet {
	size := bs.size
	if other.size < size {
		size = other.size
	}

	result := New(size)
	for i := 0; i < len(result.bits); i++ {
		if i < len(bs.bits) && i < len(other.bits) {
			result.bits[i] = bs.bits[i] & other.bits[i]
		}
	}
	return result
}

// Or performs a bitwise OR with another BitSet
func (bs *BitSet) Or(other *BitSet) *BitSet {
	size := bs.size
	if other.size > size {
		size = other.size
	}

	result := New(size)
	for i := 0; i < len(result.bits); i++ {
		var val uint64
		if i < len(bs.bits) {
			val |= bs.bits[i]
		}
		if i < len(other.bits) {
			val |= other.bits[i]
		}
		result.bits[i] = val
	}
	return result
}

// Xor performs a bitwise XOR with another BitSet
func (bs *BitSet) Xor(other *BitSet) *BitSet {
	size := bs.size
	if other.size > size {
		size = other.size
	}

	result := New(size)
	for i := 0; i < len(result.bits); i++ {
		var val uint64
		if i < len(bs.bits) {
			val ^= bs.bits[i]
		}
		if i < len(other.bits) {
			val ^= other.bits[i]
		}
		result.bits[i] = val
	}
	return result
}

// ToBools converts the BitSet to a boolean array
func (bs *BitSet) ToBools() []bool {
	bools := make([]bool, bs.size)
	for i := 0; i < bs.size; i++ {
		bools[i] = bs.Get(i)
	}
	return bools
}

// ToBytes converts the BitSet to a byte array and its bit length
func (bs *BitSet) ToBytes() []byte {
	byteLen := (bs.size + 7) / 8
	bytes := make([]byte, byteLen)

	for i := 0; i < bs.size; i++ {
		bytePos, bitPos := len(bytes)-1-i/8, i%8
		bytes[bytePos] |= byte(bs.GetInt(i) << bitPos)
	}

	return bytes
}

// ToInt converts the BitSet to an integer value and its bit length
// Note: This will only work correctly for BitSets that fit in an int
func (bs *BitSet) ToInt() int {
	if bs.size > 64 {
		panic("BitSet too large to convert to int")
	}

	var val int
	for i := 0; i < bs.size; i++ {
		val |= bs.GetInt(i) << i
	}

	return val
}

// String returns a string representation of the BitSet
func (bs *BitSet) ToString() string {
	// "0" for false, "1" for true
	bits := make([]byte, bs.size)
	for i := 0; i < bs.size; i++ {
		if bs.Get(bs.size - i - 1) {
			bits[i] = '1'
		} else {
			bits[i] = '0'
		}
	}
	return string(bits)
}

// String returns a string representation of the BitSet
func (bs *BitSet) String() string {
	return fmt.Sprintf("BitSet(%d)[0x%X]", bs.size, bs.ToBytes())
}

// Count returns the number of bits set to 1
func (bs *BitSet) Count() int {
	count := 0
	for _, word := range bs.bits {
		count += bits.OnesCount64(word)
	}
	return count
}

// Equal checks if two BitSets have the same bits
func Equal(one *BitSet, other *BitSet) bool {
	if one.size != other.size {
		return false
	}

	// Compare each word
	for i := 0; i < len(one.bits); i++ {
		if one.bits[i] != other.bits[i] {
			return false
		}
	}

	return true
}

// Equal checks if two BitSets have the same bits
func (bs *BitSet) Equals(other *BitSet) bool {
	return Equal(bs, other)
}

// Concat concatenates multiple BitSets and returns a new BitSet
func Concat(sets ...*BitSet) *BitSet {
	// Calculate total size
	newSize := 0
	for _, other := range sets {
		newSize += other.size
	}

	result := New(newSize)

	// Copy bits from other BitSets
	offset := 0
	for _, other := range sets {
		for i := 0; i < other.size; i++ {
			result.Set(offset+i, other.Get(i))
		}
		offset += other.size
	}

	return result
}

// Equal checks if two BitSets have the same bits
func (bs *BitSet) Append(other *BitSet) *BitSet {
	return Concat(bs, other)
}
