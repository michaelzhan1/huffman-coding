package compress

import "errors"

var (
	errBuildTree      = errors.New("failed to build compression tree")
	errEmptyFile      = errors.New("file is empty")
	errCompressHeader = errors.New("failed to write header to compressed file")
	errCompress       = errors.New("failed to compress file")
)
