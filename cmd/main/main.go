package main

import (
	"fmt"

	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	c, err := compress.NewCompressor("Makefile")
	if err != nil {
		panic(err)
	}

	fmt.Println(c)
	// fmt.Println(tree.GetCode())

	fmt.Println(c.GetFreq())
	c.Compress("test.huff")

	// _, err := compress.NewDecompressor("test.huff")
	// if err != nil {
	// 	panic(err)
	// }
}
