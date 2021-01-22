package jsonnetmod

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

func versionGreaterThan(v string, w string) bool {
	if w == "" {
		return false
	}
	return semver.Compare(v, w) > 0
}

func isSubDirFor(targetpath string, root string) bool {
	targetpath = targetpath + "/"
	root = root + "/"
	return strings.HasPrefix(targetpath, root)
}

func subPath(pkg string, importPath string) (string, error) {
	if isSubDirFor(importPath, pkg) {
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
