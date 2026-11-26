package main

import (
	"fmt"
	"os"

	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: ./hcomp <file> <output>")
		return
	}

	fname := os.Args[1]
	out := os.Args[2]

	c, err := compress.NewCompressor(fname)
	if err != nil {
		fmt.Printf("failed to compress file: %v\n", err)
		return
	}

	err = c.Compress(out)
	if err != nil {
		fmt.Printf("failed to compress file: %v\n", err)
		return
	}
}
