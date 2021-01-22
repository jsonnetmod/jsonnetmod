package jsonnetmod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/vcs"
)

func NewModResolver(resolved map[string]bool) *ModResolver {
	return &ModResolver{resolved: resolved}
}

type ModResolver struct {
	resolved map[string]bool
	mods     map[string]*Mod
}

func (m *ModResolver) Resolve(ctx context.Context, importPath string) (string, error) {
	for pkg := range m.resolved {
		if pathPrefix(importPath, pkg) {
			return pkg, nil
		}
	}

	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("resolve %s", importPath))

	repo, err := m.resolve(importPath)
	if err != nil {
		return "", err
	}

	if m.resolved == nil {
		m.resolved = map[string]bool{}
	}
	m.resolved[repo] = true

	return repo, nil

}

func (m *ModResolver) Get(ctx context.Context, importPath string, version string) (*Mod, error) {
	pkg, err := m.Resolve(ctx, importPath)
	if err != nil {
		return nil, err
	}
	return m.get(ctx, pkg, version)
}

func (m *ModResolver) get(ctx context.Context, pkg string, version string) (*Mod, error) {
	if m.mods == nil {
		m.mods = map[string]*Mod{}
	}

	if version == "" {
		version = "latest"
	}

	id := pkg + "@" + version

	if info, ok := m.mods[id]; ok {
		return info, nil
	}

	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("get %s", id))

	info, err := m.download(pkg, version)
	if err != nil {
		return nil, err
	}

	mod := &Mod{}
	mod.Module = info.Path
	mod.Version = info.Version
	mod.Dir = info.Dir
	mod.Sum = info.Sum

	m.mods[id] = mod

	return mod, nil
}

func (m *ModResolver) resolve(importPath string) (string, error) {
	r, err := vcs.RepoRootForImportPath(importPath, true)
	if err != nil {
		return "", errors.Wrapf(err, "resolve `%s` failed", importPath)
	}
	return r.Root, nil
}

func (m *ModResolver) download(pkg string, version string) (*GoModInfo, error) {
	buf := bytes.NewBuffer(nil)

	cmd := exec.Command("go", "mod", "download", "-json", pkg+"@"+version)
	cmd.Env = os.Environ()

	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	info := &GoModInfo{}
	if err := json.NewDecoder(buf).Decode(info); err != nil {
		return nil, err
	}
	return info, nil
}

type GoModInfo struct {
	Path    string
	Version string
	Dir     string
	Sum     string
}
