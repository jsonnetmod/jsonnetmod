package jsonnetmod

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jsonnetmod/jsonnetmod/pkg/util"
)

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

	jpath   string
	replace *replaceRule
}

type replaceRule struct {
	from string
	to   string
}

func (i ImportPath) WithReplace(from string, to string) *ImportPath {
	i.replace = &replaceRule{
		from: from,
		to:   to,
	}
	return &i
}

func (i *ImportPath) SetJPath(jpath string) {
	i.jpath = jpath
}

func (i *ImportPath) SymlinkOrTouchImportStub(ctx context.Context) error {
	if i.jpath != "" {
		if ok, m := i.shouldUseImportStub(); ok {
			return util.WriteFile(
				filepath.Join(i.jpath, i.replace.from),
				[]byte(fmt.Sprintf(`(%s '%s')`, m, filepath.Join(i.Module, i.SubPath))),
			)
		}

		if i.replace != nil && i.replace.from != i.replace.to {
			// replace should link module dir
			return i.symlink(
				ctx,
				filepath.Join(i.Dir, filepath.Dir(i.SubPath)),
				filepath.Join(i.jpath, filepath.Dir(i.ImportPath())),
			)
		}

		// should link repo
		return i.symlink(
			ctx,
			i.RepoDir(),
			filepath.Join(i.jpath, i.Repo),
		)
	}

	return nil
}

func (i *ImportPath) symlink(ctx context.Context, from string, to string) error {
	return util.Symlink(from, to)
}

func (i *ImportPath) ImportPath() string {
	if i.replace != nil && i.replace.from != i.replace.to {
		return i.replace.from + filepath.Join(i.Module, i.SubPath)[len(i.replace.to):]
	}
	return filepath.Join(i.Module, i.SubPath)
}

func (i *ImportPath) ResolvedImportPath() string {
	if i.jpath != "" {
		// return import stub
		if ok, _ := i.shouldUseImportStub(); ok {
			return filepath.Join(i.jpath, i.replace.from)
		}
		// return path linked under jpath
		return filepath.Join(i.jpath, i.ImportPath())
	}
	return filepath.Join(i.Dir, i.SubPath)
}

func (i *ImportPath) shouldUseImportStub() (ok bool, importMethod string) {
	if i.replace != nil && filepath.Ext(i.replace.to) != "" {
		ext := filepath.Ext(i.replace.from)

		if ext == "" {
			return false, ""
		}

		switch ext {
		case ".json", ".yaml", ".yml", ".jsonnet", ".libsonnet":
			return true, "import"
		default:
			return false, ""
		}
	}

	return false, ""
}

func (i *ImportPath) RepoDir() string {
	if i.Repo == i.Module {
		return i.Dir
	}
	rel, _ := subPath(i.Repo, i.Module)
	return i.Dir[0 : len(i.Dir)-len("/"+rel)]
}
