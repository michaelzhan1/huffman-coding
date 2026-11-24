package main

import (
	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	// tree, err := compress.NewCompressionTree("Makefile")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(tree)
	// fmt.Println(tree.GetCode())

	// tree.Compress("test.huff")

	_, err := compress.NewDecompressor("test.huff")
	if err != nil {
		panic(err)
	}
}
