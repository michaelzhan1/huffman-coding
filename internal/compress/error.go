package compress

import "errors"

var (
	errBuildTree = errors.New("failed to build compression tree")
	errEmptyFile = errors.New("file is empty")
)
