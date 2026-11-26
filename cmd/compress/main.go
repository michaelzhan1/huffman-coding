package main

import (
	"fmt"
	"os"

	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: ./hcomp <file> <output>")
		os.Exit(1)
	}

	fname := os.Args[1]
	out := os.Args[2]

	c, err := compress.NewCompressor(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to compress file: %v\n", err)
		os.Exit(1)
	}

	err = c.Compress(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to compress file: %v\n", err)
		os.Exit(1)
	}
}
