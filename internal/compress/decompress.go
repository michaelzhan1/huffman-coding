package compress

// import (
// 	"encoding/binary"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"sort"

// 	"github.com/michaelzhan1/huffman-coding/internal/stream"
// 	"github.com/michaelzhan1/huffman-coding/internal/util"
// )

// // Decompressor contains the information needed to decompress a file
// type Decompressor struct {
// 	tree      *huffTree
// 	fname     string
// 	totalSize uint64
// 	startByte int
// }

// func buildDecompressionTree(code map[byte]*Code) * {
// 	bytes := make([]byte, len(code))
// 	for b := range code {
// 		bytes = append(bytes, b)
// 	}

// 	// sort by increasing length
// 	sort.Slice(bytes, func(i, j int) bool { return code[bytes[i]].Length < code[bytes[j]].Length })

// 	// build tree
// 	// TODO: check correctness
// 	root := &decompressNode{}
// 	for _, b := range bytes {
// 		cur := root
// 		for i := range code[b].Length {
// 			if code[b].Bits[i/8]&(1<<(7-i%8)) == 1 {
// 				if cur.left == nil {
// 					cur.left = &decompressNode{}
// 				}
// 				cur = cur.left
// 			} else {
// 				if cur.right == nil {
// 					cur.right = &decompressNode{}
// 				}
// 				cur = cur.right
// 			}
// 		}

// 		cur.value = b
// 	}

// 	return root
// }

// // NewDecompressor returns a new Decompressor
// func NewDecompressor(fname string) (*Decompressor, error) {
// 	reader, err := stream.GetReader(fname)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer reader.Close()

// 	startByte := 0

// 	// read alphabet size (2 bytes)
// 	var alphabetCount uint16
// 	if err := binary.Read(reader, binary.BigEndian, &alphabetCount); err != nil {
// 		return nil, err
// 	}
// 	fmt.Printf("Header length: %d\n", alphabetCount)
// 	startByte += 2

// 	// read total size of content (8 bytes)
// 	var totalSize uint64
// 	if err := binary.Read(reader, binary.BigEndian, &totalSize); err != nil {
// 		return nil, err
// 	}
// 	fmt.Printf("Total size: %d\n", totalSize)
// 	startByte += 8

// 	// read alphabet ((2+(code.length+7)/8)n bytes)
// 	code := map[byte]*Code{}
// 	for range alphabetCount {
// 		// character (1 byte)
// 		b, err := reader.ReadByte()
// 		if err != nil {
// 			return nil, fmt.Errorf("%w: failed to read code character byte", errDecompress)
// 		}

// 		// code length (1 byte)
// 		lengthB, err := reader.ReadByte()
// 		if err != nil {
// 			return nil, fmt.Errorf("%w: failed to read code length byte", errDecompress)
// 		}
// 		length := int(lengthB)

// 		// code
// 		bits := [32]byte{}
// 		if _, err := io.ReadFull(reader, bits[:(length+7)/8]); err != nil {
// 			return nil, fmt.Errorf("%w: failed to read huffman code", errDecompress)
// 		}

// 		code[b] = &Code{
// 			Bits:   bits,
// 			Length: int(length),
// 		}
// 		startByte += 2 + (length+7)/8
// 	}

// 	tree := buildDecompressionTree(code)

// 	return &Decompressor{
// 		fname:     fname,
// 		totalSize: totalSize,
// 		startByte: startByte,
// 		root:      tree,
// 	}, nil
// }

// func (dec *Decompressor) Decompress(out string) error {
// 	sameFile, err := util.IsSameFile(dec.fname, out)
// 	if err != nil {
// 		return err
// 	}
// 	if sameFile {
// 		return fmt.Errorf("%w: input and output refer to the same file", errInvalidOutput)
// 	}

// 	// input reader
// 	reader, err := stream.GetReader(dec.fname)
// 	if err != nil {
// 		return err
// 	}
// 	defer reader.Close()

// 	// output writer TODO: move output writer after any input checks
// 	writer, err := stream.GetWriter(out)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		if closeErr := writer.Close(); closeErr != nil {
// 			err = errors.Join(err, closeErr)
// 		}
// 	}()

// 	// TODO: rebuild the tree in NewDecompressor

// 	// final flush
// 	if err := writer.Flush(); err != nil {
// 		return err
// 	}

// 	return nil
// }
