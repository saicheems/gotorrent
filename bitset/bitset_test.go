package bitset

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	b := New(8)
	b.Set(0)
	assert.Equal(t, []byte{128}, b.Bytes(), "they should be equal")
	b.Clear(0)
	assert.Equal(t, []byte{0}, b.Bytes(), "they should be equal")
	b.Set(0)
	b.Set(1)
	assert.Equal(t, []byte{192}, b.Bytes(), "they should be equal")
	b.Clear(0)
	assert.Equal(t, []byte{64}, b.Bytes(), "they should be equal")
	b.Set(7)
	assert.Equal(t, []byte{65}, b.Bytes(), "they should be equal")

	b = New(9)
	b.Set(8)
	assert.Equal(t, []byte{0, 128}, b.Bytes(), "they should be equal")

	assert.Equal(t, true, b.Check(8), "they should be equal")
	assert.Equal(t, false, b.Check(7), "they should be equal")

	assert.Equal(t, 0, b.FirstZeroBit(), "they should be equal")
	b.Set(0)
	assert.Equal(t, 1, b.FirstZeroBit(), "they should be equal")
	b.Set(1)
	fmt.Println(b.Bytes())
	assert.Equal(t, 2, b.FirstZeroBit(), "they should be equal")
}

func TestCheck(t *testing.T) {
	b := New(9)
	b.Set(8)
	assert.Equal(t, true, b.Check(8), "they should be equal")
}
