package main

import (
	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	c, err := compress.NewCompressor("internal/compress/decompress.go")
	if err != nil {
		panic(err)
	}
	err = c.Compress("test.huff")
	if err != nil {
		panic(err)
	}

	d, err := compress.NewDecompressor("test.huff")
	if err != nil {
		panic(err)
	}

	err = d.Decompress("decomp2.go")
	if err != nil {
		panic(err)
	}
}
