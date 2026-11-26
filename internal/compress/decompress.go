package compress

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/michaelzhan1/huffman-coding/internal/stream"
	"github.com/michaelzhan1/huffman-coding/internal/util"
)

// Decompressor contains the information needed to decompress a file
type Decompressor struct {
	tree      *huffTree
	fname     string
	totalBits uint64
	startByte int
}

// NewDecompressor returns a new Decompressor
func NewDecompressor(fname string) (*Decompressor, error) {
	reader, err := stream.GetReader(fname)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	startByte := 0

	// read alphabet size (2 bytes)
	var alphabetCount uint16
	if err := binary.Read(reader, binary.BigEndian, &alphabetCount); err != nil {
		return nil, err
	}
	startByte += 2

	// read total bits written (8 bytes)
	var totalBits uint64
	if err := binary.Read(reader, binary.BigEndian, &totalBits); err != nil {
		return nil, err
	}
	startByte += 8

	// read alphabet ((2+(codes.length+7)/8)n bytes)
	codes := map[byte]*Code{}
	for range alphabetCount {
		// character (1 byte)
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("%w: failed to read code character byte", errDecompress)
		}

		// code length (1 byte)
		lengthB, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("%w: failed to read code length byte", errDecompress)
		}
		length := int(lengthB)

		// code
		bits := [32]byte{}
		if _, err := io.ReadFull(reader, bits[:(length+7)/8]); err != nil {
			return nil, fmt.Errorf("%w: failed to read huffman code", errDecompress)
		}

		codes[b] = &Code{
			Bits:   bits,
			Length: int(length),
		}
		startByte += 2 + (length+7)/8
	}

	tree := treeFromCode(codes)

	return &Decompressor{
		tree:      tree,
		fname:     fname,
		totalBits: totalBits,
		startByte: startByte,
	}, nil
}

func (d *Decompressor) Decompress(out string) error {
	sameFile, err := util.IsSameFile(d.fname, out)
	if err != nil {
		return err
	}
	if sameFile {
		return fmt.Errorf("%w: input and output refer to the same file", errInvalidOutput)
	}

	// input reader
	reader, err := stream.GetReader(d.fname)
	if err != nil {
		return err
	}
	defer reader.Close()

	// output writer
	writer, err := stream.GetWriter(out)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	// jump to start
	reader.Discard(d.startByte)

	// decode
	decode := d.tree.newDecoder()
	bitsRead := uint64(0)

	var b byte
	for bitsRead < d.totalBits {
		if bitsRead%8 == 0 {
			b, err = reader.ReadByte()

			if err == io.EOF {
				break
			}
			if err != nil {
				return nil
			}
		}

		bit := (b >> (7 - (bitsRead % 8))) & 1
		value, found, err := decode(bit == 1)
		if err != nil {
			return err
		}

		if found {
			if err := writer.WriteByte(value); err != nil {
				return err
			}
		}
		bitsRead++
	}

	if bitsRead != d.totalBits {
		return fmt.Errorf("%w: bit read mismatch. got %d, want %d", errDecompress, bitsRead, d.totalBits)
	}

	// final flush
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}
