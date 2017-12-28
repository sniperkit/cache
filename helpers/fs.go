package helpers

import (
	"os"
	"path/filepath"
)

func FileExistsFn(basepath, filename string) bool {
	if _, err := os.Stat(filepath.Join(basepath, filename)); os.IsNotExist(err) {
		return false
	}
	return true
}

func EnsureDir(path string) {
	d, err := os.Open(path)
	if err != nil {
		os.MkdirAll(path, os.FileMode(0755))
	}
	d.Close()
}
