package jsonnetmod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/vcs"
)

func NewModCache() *ModCache {
	return &ModCache{
		mods:           map[string]*Mod{},
		replace:        map[PathIdentity]PathIdentity{},
		moduleVersions: map[string]string{},
	}
}

type ModCache struct {
	replace map[PathIdentity]PathIdentity
	// { [<module>@<version>]: *Mod }
	mods map[string]*Mod
	// { [<module>]:latest-version }
	moduleVersions map[string]string
}

func (c *ModCache) LookupReplace(importPath string, version string) (matched PathIdentity, replace PathIdentity, exists bool) {
	for _, path := range paths(importPath) {
		for _, p := range []PathIdentity{
			{Path: path, Version: ""},
			{Path: path, Version: "latest"},
			{Path: path, Version: version},
		} {
			if rp, ok := c.replace[p]; ok {
				return p, rp, true
			}
		}
	}

	return PathIdentity{}, PathIdentity{}, false
}

func (c *ModCache) Set(mod *Mod) {
	id := mod.Module + "@" + mod.Version

	c.mods[id] = mod

	c.SetModuleVersion(mod.Module, mod.Version)

	for module, r := range mod.Require {
		c.SetModuleVersion(module, r.Version)
	}

	for k, v := range mod.Replace {
		if _, ok := c.replace[k]; !ok {
			c.replace[k] = v
		}
	}
}

func (c *ModCache) SetModuleVersion(module string, version string) {
	if v, ok := c.moduleVersions[module]; ok {
		if v == "" {
			c.moduleVersions[module] = version
		} else if versionGreaterThan(version, v) {
			c.moduleVersions[module] = version
		}
	} else {
		c.moduleVersions[module] = version
	}
}

func (c *ModCache) ModuleVersion(repo string) (version string) {
	if v, ok := c.moduleVersions[repo]; ok {
		return v
	}
	return ""
}

type VersionFixer = func(repo string, version string) string

func (c *ModCache) Get(ctx context.Context, importPath string, version string, fixVersion VersionFixer) (*Mod, error) {
	module, err := c.repoRoot(ctx, importPath)
	if err != nil {
		return nil, err
	}

	if fixVersion != nil {
		version = fixVersion(module, version)
	}

	if OptsFromContext(ctx).Upgrade {
		version = "upgrade"
	}

	if version == "" {
		version = "latest"
	}

	if mod, ok := c.mods[module+"@"+version]; ok {
		if mod.Resolved() {
			return mod, nil
		}
	}

	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("get %s@%s", module, version))

	mod, err := c.download(module, version)
	if err != nil {
		return nil, err
	}

	if err := LoadModInfo(mod); err != nil {
		return nil, err
	}

	c.Set(mod)

	return mod, nil
}

func (c *ModCache) repoRoot(ctx context.Context, importPath string) (string, error) {
	for _, p := range paths(importPath) {
		if _, ok := c.moduleVersions[p]; ok {
			return p, nil
		}
	}

	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("resolve %s", importPath))

	r, err := vcs.RepoRootForImportPath(importPath, true)
	if err != nil {
		return "", errors.Wrapf(err, "resolve `%s` failed", importPath)
	}

	repo := r.Root

	// try resolve sub modules
	if importPath != repo {
		d := importPath

		for d != repo {
			if mod, err := c.download(d, "upgrade"); err == nil {
				repo = mod.Module
				break
			}

			d = filepath.Join(d, "../")
		}
	}

	c.SetModuleVersion(repo, "")

	return repo, nil
}

func paths(path string) []string {
	paths := make([]string, 0)

	d := path

	for {
		paths = append(paths, d)

		if !strings.Contains(d, "/") {
			break
		}

		d = filepath.Join(d, "../")
	}

	return paths
}

func (ModCache) download(pkg string, version string) (*Mod, error) {
	buf := bytes.NewBuffer(nil)

	cmd := exec.Command("go", "mod", "download", "-json", pkg+"@"+version)
	cmd.Env = os.Environ()

	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	info := &struct {
		Path    string
		Version string
		Error   string `json:",omitempty"`
		Dir     string `json:",omitempty"`
		Sum     string `json:",omitempty"`
	}{}
	if err := json.NewDecoder(buf).Decode(info); err != nil {
		return nil, err
	}

	if info.Error != "" {
		return nil, errors.New(info.Error)
	}

	mod := &Mod{}

	mod.Module = info.Path
	mod.Version = info.Version
	mod.Dir = info.Dir
	mod.Sum = info.Sum

	return mod, nil
}
