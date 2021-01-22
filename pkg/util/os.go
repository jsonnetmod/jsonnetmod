package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func WriteFile(filename string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, os.ModePerm)
}

func Symlink(from, to string) error {
	if err := os.RemoveAll(to); err != nil {
		return err
	}
	// make sure parent created
	if err := os.MkdirAll(filepath.Dir(to), 0777); err != nil {
		return err
	}
	if err := os.Symlink(from, to); err != nil {
		return err
	}
	return nil
}
