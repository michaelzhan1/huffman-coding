package stream

import (
	"bufio"
	"io"
	"os"
)

// BufferedReadCloser is a struct that implements io.Reader and io.Closer.
// It is used to stream bytes from a file without loading the file entirely into memory.
type BufferedReadCloser struct {
	*bufio.Reader
	io.Closer
}

// GetReader returns a BufferedReadCloser from a file.
func GetReader(fname string) (*BufferedReadCloser, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	return &BufferedReadCloser{
		Reader: bufio.NewReader(file),
		Closer: file,
	}, nil
}

type BufferedWriteCloser struct {
	*bufio.Writer
	io.Closer
}

func GetWriter(fname string) (*BufferedWriteCloser, error) {
	file, err := os.Create(fname)
	if err != nil {
		return nil, err
	}

	return &BufferedWriteCloser{
		Writer: bufio.NewWriter(file),
		Closer: file,
	}, nil
}
