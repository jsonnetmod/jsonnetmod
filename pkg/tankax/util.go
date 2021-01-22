package tankax

import (
	"path/filepath"
	"strings"
)

func lastDirname(file string) string {
	paths := strings.Split(file, string(filepath.Separator))
	return paths[len(paths)-2]
}
