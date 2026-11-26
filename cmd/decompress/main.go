package main

import (
	"fmt"
	"os"

	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: ./hdecomp <file> <output>")
		return
	}

	fname := os.Args[1]
	out := os.Args[2]

	d, err := compress.NewDecompressor(fname)
	if err != nil {
		fmt.Printf("failed to decompress file: %v\n", err)
		return
	}

	err = d.Decompress(out)
	if err != nil {
		fmt.Printf("failed to decompress file: %v\n", err)
		return
	}
}
