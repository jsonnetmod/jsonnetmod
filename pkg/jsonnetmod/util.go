package jsonnetmod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func symlink(from, to string) error {
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

func pathPrefix(targetpath string, root string) bool {
	targetpath = targetpath + "/"
	root = root + "/"
	return strings.HasPrefix(targetpath, root)
}

func subPath(pkg string, importPath string) (string, error) {
	if pathPrefix(importPath, pkg) {
		if len(importPath) > len(pkg)+1 {
			return importPath[len(pkg)+1:], nil
		}
		return "", nil
	}
	return "", fmt.Errorf("%s is not sub path of %s", importPath, pkg)
}

func replaceImportPath(to string, from string, importPath string) string {
	if from == importPath {
		return to
	}
	s, _ := subPath(from, importPath)
	return filepath.Join(to, s)
}
