package compress

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/michaelzhan1/huffman-coding/internal/stream"
	"github.com/michaelzhan1/huffman-coding/internal/util"
	"github.com/michaelzhan1/priority-queue-go/priorityqueue"
)

// compressNode is a node in the compression tree
type compressNode struct {
	value  byte
	weight int
	left   *compressNode
	right  *compressNode
}

// CompressionTree is a compression tree
type CompressionTree struct {
	root  *compressNode
	freq  map[byte]int
	code  map[byte]*Code
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
	pq := priorityqueue.NewPriorityQueue(func(a, b *compressNode) bool { return a.weight < b.weight })
	for b, count := range freq {
		pq.Push(&compressNode{
			value:  b,
			weight: count,
		})
	}

	var root *compressNode
	for !pq.IsEmpty() {
		n1, ok := pq.Pop()
		if !ok {
			return nil, fmt.Errorf("%w: failed to pop from priority queue", errBuildTree)
		}

		if pq.IsEmpty() {
			root = n1
			break
		}

		n2, ok := pq.Pop()
		if !ok {
			return nil, fmt.Errorf("%w: failed to pop from priority queue", errBuildTree)
		}

		parent := &compressNode{
			weight: n1.weight + n2.weight,
			left:   n2,
			right:  n1,
		}

		pq.Push(parent)
	}

	if root == nil {
		return nil, fmt.Errorf("%w: failed to build root for tree", errBuildTree)
	}

	tree := &CompressionTree{
		root:  root,
		freq:  freq,
		fname: fname,
	}

	tree.code = tree.codeTable()
	return tree, nil
}

// String returns the string representation of the tree
func (tree *CompressionTree) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Frequencies:\n%v\n\n", tree.freq)

	var visit func(*compressNode, []bool) string
	visit = func(v *compressNode, pipeDepth []bool) string {
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

// codeTable constructs and returns the Huffman code table for the compression tree
func (tree *CompressionTree) codeTable() map[byte]*Code {
	res := map[byte]*Code{}

	var visit func(*compressNode, []bool)
	visit = func(v *compressNode, path []bool) {
		if v == nil {
			return
		}

		// leaf
		if v.left == nil && v.right == nil {
			code := &Code{
				Bits:   [MaxCodeBytes]byte{},
				Length: len(path),
			}

			idx := 0 // just use i?
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

// GetFreq returns the frequency table of the compression tree
func (tree *CompressionTree) GetFreq() map[byte]int {
	return tree.freq
}

// GetCode returns the Huffman code table
func (tree *CompressionTree) GetCode() map[byte]*Code {
	return tree.code
}

// Compress re-reads the file and compresses based on the Huffman coding
func (tree *CompressionTree) Compress(out string) (err error) {
	sameFile, err := util.IsSameFile(tree.fname, out)
	if err != nil {
		return err
	}
	if sameFile {
		return fmt.Errorf("%w: input and output refer to the same file", errInvalidOutput)
	}

	writer, err := stream.GetWriter(out)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	// write expected size of alphabet (2 bytes)
	if len(tree.code) > 256 {
		return fmt.Errorf("%w: unexpectedly found more than 256 alphabet characters, got %d", errCompressHeader, len(tree.code))
	}
	if err := binary.Write(writer, binary.BigEndian, uint16(len(tree.code))); err != nil {
		return err
	}

	// write size (in bits) of read (8 bytes)
	var totalSize uint64
	for b, code := range tree.code {
		count, found := tree.freq[b]
		if !found {
			return fmt.Errorf("%w: did not find count for character %q", errCompressHeader, string(b))
		}
		totalSize += uint64(code.Length) * uint64(count)
	}
	if err := binary.Write(writer, binary.BigEndian, totalSize); err != nil {
		return err
	}

	// write alphabet ((2+(code.length+7)/8)n bytes)
	for b, code := range tree.code {
		// character (1 byte)
		if err := binary.Write(writer, binary.BigEndian, b); err != nil {
			return err
		}

		// code length (1 byte)
		if code.Length > 255 || code.Length <= 0 {
			return fmt.Errorf("%w: unexpected code length outside of [1, 255] for byte %q, got %d", errCompressHeader, string(b), code.Length)
		}
		if err := writer.WriteByte(uint8(code.Length)); err != nil {
			return err
		}

		// code ((code.length+7)/8 bytes)
		if _, err := writer.Write(code.Bits[:(code.Length+7)/8]); err != nil {
			return err
		}
	}

	// compress file contents
	reader, err := stream.GetReader(tree.fname)
	if err != nil {
		return err
	}
	defer reader.Close()

	var bitsWritten uint64
	var buf byte
	rem := 8
	for {
		b, err := reader.ReadByte()

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		code, found := tree.code[b]
		if !found {
			return fmt.Errorf("%w: did not find code for %q", errCompress, string(b))
		}

		for i := range code.Length {
			bit := (code.Bits[i/8] >> (7 - i%8)) & 1
			buf |= bit << (rem - 1)
			rem--

			if rem == 0 {
				rem = 8
				if err := writer.WriteByte(buf); err != nil {
					return err
				}
				buf = 0
			}
		}
		bitsWritten += uint64(code.Length)
	}
	if rem < 8 {
		if err := writer.WriteByte(buf); err != nil {
			return err
		}
	}

	if bitsWritten != totalSize {
		return fmt.Errorf("%w: expected %d bits, wrote %d", errCompress, totalSize, bitsWritten)
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}
