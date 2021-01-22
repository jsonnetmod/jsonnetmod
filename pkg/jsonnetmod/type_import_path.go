package jsonnetmod

import "path/filepath"

func ImportPathFor(mod *Mod, importPath string) *ImportPath {
	i := &ImportPath{}
	i.Mod = mod

	if importPath != "" {
		i.SubPath, _ = subPath(i.Module, importPath)
	}

	return i
}

type ImportPath struct {
	*Mod
	SubPath string
}

func (i *ImportPath) FullPath() string {
	return filepath.Join(i.Dir, i.SubPath)
}
