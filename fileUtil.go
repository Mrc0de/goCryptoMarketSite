package main

import (
	"errors"
	"os"
)

// Misc
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, errors.New("FileNotFound") }
	return false, err
}

