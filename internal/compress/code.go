package compress

import (
	"fmt"
	"strings"
)

// Code represents a huffman Code. byte-by-byte reading to generate the tree means 256 unique values,
// which means 255-bit = 32-byte max Code
type Code struct {
	Bits   [MaxCodeBytes]byte
	Length int
}

// String returns a string representation of the Huffman code
func (code Code) String() string {
	var sb strings.Builder

	for i := 0; i*8 < code.Length; i++ {
		if i*8+8 <= code.Length {
			fmt.Fprintf(&sb, "%08b", code.Bits[i])
		} else {
			rem := code.Length - i*8
			fmt.Fprintf(&sb, "%0*b", rem, code.Bits[i]>>(8-rem))
		}
	}

	return sb.String()
}
