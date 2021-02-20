package jsonnetmod

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/octohelm/jsonnetmod/pkg/util"
)

func VModFor(root string) *VMod {
	vm := &VMod{
		cache: NewModCache(),
	}

	if !filepath.IsAbs(root) {
		cwd, _ := os.Getwd()
		root = filepath.Join(cwd, root)
	}

	mod, err := ModFromDir(root)
	if err != nil {
		panic(err)
	}

	vm.Mod = mod
	vm.cache.Set(mod)

	return vm
}

type VMod struct {
	*Mod
	cache *ModCache
}

func (v *VMod) MakeVM(ctx context.Context) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	vm.Importer(NewImporter(ctx, v))
	return vm
}

func (v *VMod) ListJsonnet(fromPath string) ([]string, error) {
	files := make([]string, 0)

	// repoRoot *.jsonnet & *.libsonnet
	start := filepath.Join(v.Dir, fromPath)

	err := filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
		if path == start {
			return nil
		}

		// skip vendor
		if v.JPath != "" {
			if isSubDirFor(path, filepath.Join(v.Dir, v.
				JPath)) {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(path, modJsonnetFile)); err == nil {
				return filepath.SkipDir
			}
			// skip sub root
			return nil
		}

		ext := filepath.Ext(path)

		if ext == ".jsonnet" || ext == ".libsonnet" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (v *VMod) Get(ctx context.Context, i string) error {
	if i[0] == '.' {
		return v.autoImport(ctx, i)
	}
	return v.download(ctx, i)
}

func (v *VMod) Resolve(ctx context.Context, importPath string, importedFrom string) (string, error) {
	resolvedImportPath, err := v.ResolveImportPath(ctx, v.cache, importPath, "")
	if err != nil {
		return "", err
	}

	if err := v.SetRequireFromImportPath(resolvedImportPath, !isSubDirFor(importedFrom, v.Dir)); err != nil {
		return "", err
	}

	return resolvedImportPath.FullPath(), nil
}

func (v *VMod) SetRequireFromImportPath(p *ImportPath, indirect bool) error {
	version := p.Version

	if ver := v.cache.ModuleVersion(p.Module); ver != "" {
		version = ver
	}

	v.SetRequire(p.Module, version, indirect)

	if v.JPath != "" {
		if err := util.Symlink(p.Dir, filepath.Join(v.Dir, v.JPath, p.Module)); err != nil {
			return err
		}

		for from, to := range v.cache.replace {
			if from.Path != to.Path && to.Path[0] != '.' {
				if isSubDirFor(to.Path, p.Module) {
					if err := util.Symlink(
						path.Join(p.Dir, p.SubPath),
						filepath.Join(v.Dir, v.JPath, from.Path),
					); err != nil {
						return err
					}
				}
			}
		}

		// hack k.libsonnet
		_ = util.WriteFile(path.Join(p.Dir, v.JPath, "k.libsonnet"), []byte(`import "k/main.libsonnet"`))
	}

	return WriteMod(v.Mod)
}

func (v *VMod) autoImport(ctx context.Context, fromPath string) error {
	vm := v.MakeVM(ctx)
	files, err := v.ListJsonnet(fromPath)
	if err != nil {
		return err
	}
	// to trigger importer
	if _, err := vm.FindDependencies("", files); err != nil {
		return err
	}
	return nil
}

func (v *VMod) download(ctx context.Context, importPath string) error {
	importPathAndVersion := strings.Split(importPath, "@")

	importPath, version := importPathAndVersion[0], ""
	if len(importPathAndVersion) > 1 {
		version = importPathAndVersion[1]
	}

	p, err := v.Mod.ResolveImportPath(ctx, v.cache, importPath, version)
	if err != nil {
		return err
	}

	return v.SetRequireFromImportPath(p, true)
}
