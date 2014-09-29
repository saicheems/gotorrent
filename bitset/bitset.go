// Package bitset implements a bitset structure to be used for bittorrent.
package bitset

type BitSet struct {
	length int
	data   []byte
}

// New returns a BitSet of n bits.
func New(n int) *BitSet {
	b := new(BitSet)
	b.length = n
	b.data = make([]byte, (n+7)>>3)
	return b
}

func (b *BitSet) checkRange(n int) {
	if n < 0 || n >= b.length {
		panic("index out of range")
	}
}

// Set sets the nth bit to 1.
func (b *BitSet) Set(n int) {
	b.checkRange(n)
	m := uint(n)
	b.data[n>>3] |= 1 << (7 - m%8)
}

// Clear sets the nth bit to 0.
func (b *BitSet) Clear(n int) {
	b.checkRange(n)
	m := uint(n)
	b.data[n>>3] &= ^(1 << (7 - m%8))
}

// Check returns the value of the nth bit.
func (b *BitSet) Check(n int) bool {
	b.checkRange(n)
	return (b.data[n>>3] & (1 << (7 - uint(n)%8))) > 0
}

// Bytes returns the byte representation.
func (b *BitSet) Bytes() []byte {
	return b.data
}

// FirstZeroBit returns the index of the first zeroed bit in the set.
// TODO: Mega inefficient, redo!!
func (b *BitSet) FirstZeroBit() int {
	for n, _ := range b.data {
		for i := 0; i < 8; i++ {
			if n*8+i >= b.length {
				break
			}
			if !b.Check(n*8 + i) {
				return n*8 + i
			}
		}
	}
	return -1
}
