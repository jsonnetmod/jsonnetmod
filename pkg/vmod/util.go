package vmod

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func readYAMLFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return yaml.NewDecoder(file).Decode(data)
}

func writeYAMLFile(filename string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0666); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		if os.IsExist(err) {
			file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		}
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	return yaml.NewEncoder(file).Encode(data)
}

func symlink(from, to string) error {
	if err := os.RemoveAll(to); err != nil {
		return err
	}
	// mkdir parent created
	if err := os.MkdirAll(filepath.Dir(to), 0777); err != nil {
		return err
	}
	if err := os.Symlink(from, to); err != nil {
		return err
	}
	return nil
}
