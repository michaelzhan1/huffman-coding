package util

import (
	"errors"
	"os"
)

func IsSameFile(f1, f2 string) (bool, error) {
	info1, err := os.Stat(f1)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	info2, err := os.Stat(f2)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return os.SameFile(info1, info2), nil
}
