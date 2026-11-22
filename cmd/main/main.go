package main

import (
	"fmt"

	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	tree, err := compress.NewCompressionTree("Makefile")
	if err != nil {
		panic(err)
	}

	fmt.Println(tree)

	fmt.Println(tree.CodeTable())
}
