package jsonnetmod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/vcs"
)

func NewModCache() *ModCache {
	return &ModCache{
		mods:           map[string]*Mod{},
		replace:        map[PathReplace]PathReplace{},
		moduleVersions: map[string]string{},
	}
}

type ModCache struct {
	replace map[PathReplace]PathReplace
	// { [<module>@<version>]: *Mod }
	mods map[string]*Mod
	// { [<module>]:latest-version }
	moduleVersions map[string]string
}

func (c *ModCache) LookupReplace(importPath string, version string) (matched PathReplace, replace PathReplace, exists bool) {
	for p, rp := range c.replace {
		if isSubDirFor(importPath, p.Path) {
			if version == "" || version == "latest" {
				return p, rp, true
			}
			if version == p.Version {
				return p, rp, true
			}
		}
	}
	return PathReplace{}, PathReplace{}, false
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
		version = "latest"
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

	c.SetModuleVersion(r.Root, "")

	return r.Root, nil
}

func paths(path string) []string {
	parts := strings.Split(path, "/")

	paths := make([]string, len(parts))

	prefix := ""

	for i, p := range parts {
		paths[i] = prefix + p
		prefix = paths[i] + "/"
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
