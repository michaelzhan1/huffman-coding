package compress

import (
	"fmt"
	"io"
	"strings"

	"github.com/michaelzhan1/huffman-coding/internal/stream"
	"github.com/michaelzhan1/priority-queue-go/priorityqueue"
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

	for i := 0; i*8 <= code.Length; i++ {
		if i*8+8 <= code.Length {
			fmt.Fprintf(&sb, "%08b", code.Bits[i])
		} else {
			rem := code.Length - i*8
			fmt.Fprintf(&sb, "%0*b", rem, code.Bits[i]>>(8-rem))
		}
	}

	return sb.String()
}

// node is a node in the compression tree
type node struct {
	value  byte
	weight int
	left   *node
	right  *node
}

// CompressionTree is a compression tree
type CompressionTree struct {
	root  *node
	freq  map[byte]int
	fname string
}

// NewCompressionTree returns a new compression tree
func NewCompressionTree(fname string) (*CompressionTree, error) {
	reader, err := stream.GetReader(fname)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// build frequency map
	freq := map[byte]int{}
	for {
		b, err := reader.ReadByte()

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		freq[b]++
	}

	// check for empty file
	if len(freq) == 0 {
		return nil, errEmptyFile
	}

	// build tree
	pq := priorityqueue.NewPriorityQueue(func(a, b *node) bool { return a.weight < b.weight })
	for b, count := range freq {
		pq.Push(&node{
			value:  b,
			weight: count,
		})
	}

	for !pq.IsEmpty() {
		n1, ok := pq.Pop()
		if !ok {
			return nil, fmt.Errorf("%w: failed to pop from priority queue", errBuildTree)
		}

		if pq.IsEmpty() {
			return &CompressionTree{
				root:  n1,
				freq:  freq,
				fname: fname,
			}, nil
		}

		n2, ok := pq.Pop()
		if !ok {
			return nil, fmt.Errorf("%w: failed to pop from priority queue", errBuildTree)
		}

		parent := &node{
			weight: n1.weight + n2.weight,
			left:   n2,
			right:  n1,
		}

		pq.Push(parent)
	}

	return nil, fmt.Errorf("%w: unexpectedly exited tree building loop", errBuildTree)
}

// String returns the string representation of the tree
func (tree *CompressionTree) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Frequencies:\n%v\n\n", tree.freq)

	var visit func(*node, []bool) string
	visit = func(v *node, pipeDepth []bool) string {
		var tab strings.Builder
		for _, usePipe := range pipeDepth {
			if usePipe {
				tab.WriteString("\u2502 ")
			} else {
				tab.WriteString("  ")
			}
		}

		branchPrefix := tab.String() + "\u251c\u2500"
		bendPrefix := tab.String() + "\u2514\u2500"

		if v == nil {
			return bendPrefix + "nil\n"
		}

		var value string
		if v.value > 0 {
			value = string(v.value)
		}

		pipeDepth = append(pipeDepth, true)
		left := visit(v.left, pipeDepth)
		pipeDepth[len(pipeDepth)-1] = false
		right := visit(v.right, pipeDepth)
		pipeDepth = pipeDepth[:len(pipeDepth)-1]

		var sb strings.Builder
		fmt.Fprintf(&sb, branchPrefix+"Value: %q\n", value)
		fmt.Fprintf(&sb, branchPrefix+"Weight: %d\n", v.weight)
		fmt.Fprintf(&sb, branchPrefix+"Left:\n%s", left)
		fmt.Fprintf(&sb, bendPrefix+"Right:\n%s", right)

		return sb.String()
	}

	sb.WriteString("Tree:\n")
	sb.WriteString(visit(tree.root, []bool{}))

	return sb.String()
}

// GetFreq returns the frequency table of the compression tree
func (tree *CompressionTree) GetFreq() map[byte]int {
	return tree.freq
}

// CodeTable constructs and returns the Huffman code table for the compression tree
func (tree *CompressionTree) CodeTable() map[byte]*Code {
	res := map[byte]*Code{}

	var visit func(*node, []bool)
	visit = func(v *node, path []bool) {
		if v == nil {
			return
		}

		// leaf
		if v.left == nil && v.right == nil {
			code := &Code{
				Bits:   [MaxCodeBytes]byte{},
				Length: len(path),
			}

			idx := 0
			for i, bit := range path {
				if bit {
					code.Bits[idx] |= 1 << (7 - i)
				}
				if (i+1)%8 == 0 {
					idx++
				}
			}

			res[v.value] = code
			return
		}

		// traverse
		path = append(path, false)
		visit(v.left, path)
		path[len(path)-1] = true
		visit(v.right, path)
		path = path[:len(path)-1]
	}

	visit(tree.root, []bool{})
	return res
}

// Compress re-reads the file and compresses based on the Huffman coding
func (tree *CompressionTree) Compress(out string) error {
	return nil
}
