package jsonnetmod

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jsonnetmod/jsonnetmod/pkg/jsonnetmod/modfile"

	"github.com/google/go-jsonnet"
)

func VModFor(root string) *VMod {
	vm := &VMod{
		cache: NewModCache(),
	}

	if !filepath.IsAbs(root) {
		cwd, _ := os.Getwd()
		root = filepath.Join(cwd, root)
	}

	mod := &Mod{}
	mod.Dir = root

	if _, err := mod.LoadInfo(); err != nil {
		panic(err)
	}

	vm.Mod = mod
	vm.cache.Collect(context.Background(), mod)

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
			if _, err := os.Stat(filepath.Join(path, modfile.ModFilename)); err == nil {
				return filepath.SkipDir
			}
			// skip sub repoRoot
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

	indirect := !isSubDirFor(importedFrom, v.Dir)
	if v.JPath != "" {
		indirect = isSubDirFor(importedFrom, path.Join(v.Dir, v.JPath))
	}

	if err := v.SetRequireFromImportPath(resolvedImportPath, indirect); err != nil {
		return "", err
	}

	dir := resolvedImportPath.ResolvedImportPath()

	return dir, nil
}

func (v *VMod) SetRequireFromImportPath(p *ImportPath, indirect bool) error {
	modVersion := p.ModVersion

	if mv := v.cache.RepoVersion(p.Repo); mv.Version != "" {
		modVersion = mv
	}

	v.SetRequire(p.Repo, modVersion, indirect)

	if v.JPath != "" {
		// create symlink
		p.SetJPath(filepath.Join(v.Dir, v.JPath))

		if err := p.SymlinkOrTouchImportStub(); err != nil {
			return err
		}
	}

	return modfile.WriteModFile(v.Dir, &v.ModFile)
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
