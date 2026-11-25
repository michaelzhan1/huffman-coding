package compress

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"maps"
	"strings"

	"github.com/michaelzhan1/huffman-coding/internal/stream"
	"github.com/michaelzhan1/huffman-coding/internal/util"
)

// Compressor contains information for compressing
type Compressor struct {
	tree  *huffTree
	fname string
	freq  map[byte]int // TODO: use uint64
}

// NewCompressor returns a new Compressor
func NewCompressor(fname string) (*Compressor, error) {
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
	tree, err := treeFromFreq(freq)
	if err != nil {
		return nil, err
	}

	return &Compressor{
		tree:  tree,
		fname: fname,
		freq:  freq,
	}, nil
}

// GetFreq returns the frequency map of the file
func (c *Compressor) GetFreq() map[byte]int {
	return maps.Clone(c.freq)
}

// String returns a string representation of the compressor
func (c *Compressor) String() string {
	var sb strings.Builder
	sb.WriteString(c.fname)
	sb.WriteByte('\n')
	sb.WriteString(c.tree.String())
	return sb.String()
}

// Compress re-reads the file and compresses based on the Huffman coding
func (c *Compressor) Compress(out string) (err error) {
	sameFile, err := util.IsSameFile(c.fname, out)
	if err != nil {
		return err
	}
	if sameFile {
		return fmt.Errorf("%w: input and output refer to the same file", errInvalidOutput)
	}

	codes := c.tree.getCodeMap()

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
	if len(codes) > 256 {
		return fmt.Errorf("%w: unexpectedly found more than 256 alphabet characters, got %d", errCompressHeader, len(codes))
	}
	if err := binary.Write(writer, binary.BigEndian, uint16(len(codes))); err != nil {
		return err
	}

	// write size (in bits) of read (8 bytes)
	var totalSize uint64
	for b, code := range codes {
		count, found := c.freq[b]
		if !found {
			return fmt.Errorf("%w: did not find count for character %q", errCompressHeader, string(b))
		}
		totalSize += uint64(code.Length) * uint64(count)
	}
	if err := binary.Write(writer, binary.BigEndian, totalSize); err != nil {
		return err
	}

	// write alphabet ((2+(code.length+7)/8)n bytes)
	for b, code := range codes {
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
	reader, err := stream.GetReader(c.fname)
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

		code, found := codes[b]
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
