package main

import (
	"fmt"
	"io"

	"github.com/michaelzhan1/huffman-coding/internal/compress"
)

func main() {
	reader, err := compress.GetReader("Makefile")
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	for {
		b, err := reader.ReadByte()

		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		fmt.Printf("%c", b)
	}
}
