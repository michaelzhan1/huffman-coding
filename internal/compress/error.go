package compress

import "errors"

var (
	errBuildCompressTree = errors.New("failed to build compression tree")
	errEmptyFile         = errors.New("file is empty")
	errInvalidOutput     = errors.New("invalid output file")

	errCompressHeader = errors.New("failed to write header to compressed file")
	errCompress       = errors.New("failed to compress file")

	errDecompress = errors.New("failed to decompress file")
)
