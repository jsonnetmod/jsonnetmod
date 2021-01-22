package jsonnetmod

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/pkg/errors"
)

type VModOpts struct {
	Upgrade bool
}

type contextKeyForVModOpts int

func VModOptsFromContext(ctx context.Context) VModOpts {
	if o, ok := ctx.Value(contextKeyForVModOpts(0)).(VModOpts); ok {
		return o
	}
	return VModOpts{}
}

func WithVModOpts(o VModOpts) func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKeyForVModOpts(0), o)
	}
}

func VModFor(root string) *VMod {
	vm := &VMod{}

	if !filepath.IsAbs(root) {
		cwd, _ := os.Getwd()
		root = filepath.Join(cwd, root)
	}

	vm.Dir = root

	if err := vm.Load(); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}

	resolved := map[string]bool{}

	for pkg := range vm.Require {
		resolved[pkg] = true
	}

	vm.resolver = NewModResolver(resolved)

	return vm
}

type VMod struct {
	Mod
	resolver *ModResolver
}

func (v *VMod) ListJsonnet(fromPath string) ([]string, error) {
	files := make([]string, 0)

	// resolve *.jsonnet & *.libsonnet
	start := filepath.Join(v.Dir, fromPath)

	err := filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
		if path == start {
			return nil
		}

		// skip vendor
		if v.JPath != "" {
			if pathPrefix(path, filepath.Join(v.Dir, v.JPath)) {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(path, "mod.jsonnet")); err == nil {
				return filepath.SkipDir
			}
			// skip sub mod
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
	resolvedImportPath, err := v.resolve(ctx, importPath, "")
	if err != nil {
		return "", err
	}

	indirect := !pathPrefix(importedFrom, v.Dir)

	dir := resolvedImportPath.AbsolutePath()

	// try to resolve sub mod
	mod := Mod{Dir: dir}
	if err := mod.Load(); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
	} else {
		for pkg, require := range mod.Require {
			v.SetRequire(pkg, require.Version, true)
		}
	}

	if err := v.SetRequireFromImportPath(resolvedImportPath, indirect); err != nil {
		return "", err
	}

	return resolvedImportPath.AbsolutePath(), nil
}

func (v *VMod) MakeVM(ctx context.Context) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	vm.Importer(NewImporter(ctx, v))
	return vm
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

	p, err := v.resolve(ctx, importPath, version)
	if err != nil {
		return err
	}

	return v.SetRequireFromImportPath(p, true)
}

func (v *VMod) SetRequireFromImportPath(p *ImportPath, indirect bool) error {
	v.SetRequire(p.Module, p.Version, indirect)

	if v.JPath != "" {
		if err := symlink(p.Dir, filepath.Join(v.Dir, v.JPath, p.Module)); err != nil {
			return err
		}
	}

	return v.Write()
}

func (v *VMod) resolve(ctx context.Context, importPath string, version string) (*ImportPath, error) {
	// self import '<mod.module>/dir/to/sub'
	if pathPrefix(importPath, v.Module) {
		return NewImportPath(v.Module, version, importPath).WithDir(v.Dir), nil
	}

	if matched, replace, ok := v.LookupReplace(importPath, version); ok {
		// xxx => ../xxx
		if replace.IsLocalReplace() {
			dir := filepath.Join(v.Dir, replace.Path)
			_, err := os.Stat(dir)
			if err == nil {
				return NewImportPath(matched.Path, version, importPath).WithDir(dir), nil
			}
			return nil, errors.Wrapf(err, "%s not found", dir)
		}
		// a[@latest] => b@latest
		// must strict version
		return v.resolve(WithVModOpts(VModOpts{Upgrade: false})(ctx), replaceImportPath(replace.Path, matched.Path, importPath), replace.Version)
	}

	repo, err := v.resolver.Resolve(ctx, importPath)
	if err != nil {
		return nil, err
	}

	if VModOptsFromContext(ctx).Upgrade {
		version = "latest"
	}

	if version == "" {
		if p, ok := v.Lookup(repo); ok {
			version = p.Version
		} else {
			version = "latest"
		}
	}

	info, err := v.resolver.Get(ctx, repo, version)
	if err != nil {
		return nil, err
	}

	return NewImportPath(info.Module, info.Version, importPath).WithDir(info.Dir), nil
}
