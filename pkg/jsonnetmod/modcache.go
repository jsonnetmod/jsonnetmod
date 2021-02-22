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

	"golang.org/x/tools/go/vcs"

	"github.com/octohelm/jsonnetmod/pkg/jsonnetmod/modfile"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

func NewModCache() *ModCache {
	return &ModCache{
		replace:      map[modfile.PathIdentity]pathIdentityWithMod{},
		mods:         map[string]*Mod{},
		repoVersions: map[string]string{},
	}
}

type ModCache struct {
	replace map[modfile.PathIdentity]pathIdentityWithMod
	// { [<module>@<version>]: *Mod }
	mods map[string]*Mod
	// { [<repo>]:latest-version }
	repoVersions map[string]string
}

type pathIdentityWithMod struct {
	modfile.PathIdentity
	mod *Mod
}

func (c *ModCache) LookupReplace(importPath string, version string) (matched modfile.PathIdentity, replace modfile.PathIdentity, exists bool) {
	for _, path := range paths(importPath) {
		for _, p := range []modfile.PathIdentity{
			{Path: path, Version: ""},
			{Path: path, Version: "latest"},
			{Path: path, Version: version},
		} {
			if rp, ok := c.replace[p]; ok {
				return p, rp.PathIdentity, true
			}
		}
	}

	return modfile.PathIdentity{}, modfile.PathIdentity{}, false
}

func (c *ModCache) Collect(ctx context.Context, mod *Mod) {
	id := mod.String()

	if mod.Repo == "" {
		mod.Repo = mod.Module
	}

	c.mods[id] = mod

	c.SetRepoVersion(mod.Repo, mod.Version)

	for repo, r := range mod.Require {
		c.SetRepoVersion(repo, r.Version)
	}

	for k, replaceTarget := range mod.Replace {
		if currentReplaceTarget, ok := c.replace[k]; !ok {
			c.replace[k] = pathIdentityWithMod{mod: mod, PathIdentity: replaceTarget}
		} else {
			if replaceTarget.String() != currentReplaceTarget.PathIdentity.String() {
				fmt.Printf(`
[WARNING] '%s' already replaced to 
	'%s' (using by module '%s'), but another module want to replace
	'%s' (requested by module %s)
`,
					k,
					currentReplaceTarget.PathIdentity, currentReplaceTarget.mod,
					replaceTarget, mod,
				)
			}
		}
	}
}

func (c *ModCache) SetRepoVersion(module string, version string) {
	if v, ok := c.repoVersions[module]; ok {
		if v == "" {
			c.repoVersions[module] = version
		} else if versionGreaterThan(version, v) {
			c.repoVersions[module] = version
		}
	} else {
		c.repoVersions[module] = version
	}
}

func (c *ModCache) RepoVersion(repo string) (version string) {
	if v, ok := c.repoVersions[repo]; ok {
		return v
	}
	return ""
}

type VersionFixer = func(repo string, version string) string

func (c *ModCache) Get(ctx context.Context, importPath string, version string, fixVersion VersionFixer) (*Mod, error) {
	repo, err := c.repoRoot(ctx, importPath)
	if err != nil {
		return nil, err
	}

	if fixVersion != nil {
		version = fixVersion(repo, version)
	}

	if OptsFromContext(ctx).Upgrade {
		version = "upgrade"
	}

	if version == "" {
		version = "latest"
	}

	if version != "upgrade" {
		if mod, ok := c.mods[repo+"@"+version]; ok {
			if mod.Resolved() {
				return mod, nil
			}
		}
	}

	if r, ok := c.replace[modfile.PathIdentity{Path: repo, Version: version}]; ok {
		repo, version = r.Path, r.Version
	}

	return c.get(ctx, repo, version, importPath)
}

func (c *ModCache) get(ctx context.Context, module string, version string, importPath string) (*Mod, error) {
	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("get %s@%s", module, version))

	root, err := c.download(module, version)
	if err != nil {
		return nil, err
	}

	// sub dir may as mod.
	importPaths := paths(importPath)

	c.Collect(ctx, root)

	for _, p := range importPaths {
		if p == root.Module {
			break
		}

		rel, _ := subPath(root.Module, p)

		sub := Mod{}
		sub.Repo = root.Repo
		sub.Module = p
		sub.Version = root.Version
		sub.Sum = root.Sum

		sub.Dir = filepath.Join(root.Dir, rel)

		ok, err := sub.LoadInfo()
		if err != nil {
			// if dir contains go.mod, will be empty
			if os.IsNotExist(errors.Unwrap(err)); err != nil {
				return c.get(ctx, sub.Module, version, importPath)
			}
			return nil, err
		}

		if ok {
			c.Collect(ctx, &sub)
			return &sub, nil
		}
	}

	return root, nil
}

func (c *ModCache) repoRoot(ctx context.Context, importPath string) (string, error) {
	importPaths := paths(importPath)

	for _, p := range importPaths {
		if _, ok := c.repoVersions[p]; ok {
			return p, nil
		}
	}

	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("resolve %s", importPath))

	r, err := vcs.RepoRootForImportPath(importPath, true)
	if err != nil {
		return "", errors.Wrapf(err, "resolve `%s` failed", importPath)
	}

	c.SetRepoVersion(r.Root, "")

	return r.Root, nil
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
	mod.Repo = info.Path
	mod.Version = info.Version
	mod.Dir = info.Dir
	mod.Sum = info.Sum

	return mod, nil
}
