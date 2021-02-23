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
	"github.com/octohelm/jsonnetmod/pkg/jsonnetmod/modfile"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/vcs"
)

func NewModCache() *ModCache {
	return &ModCache{
		replace:      map[modfile.PathIdentity]pathIdentityWithMod{},
		mods:         map[string]*Mod{},
		repoVersions: map[string]modfile.ModVersion{},
	}
}

type ModCache struct {
	replace map[modfile.PathIdentity]pathIdentityWithMod
	// { [<module>@<version>]: *Mod }
	mods map[string]*Mod
	// { [<repo>]:latest-version }
	repoVersions map[string]modfile.ModVersion
}

type pathIdentityWithMod struct {
	modfile.PathIdentity
	mod *Mod
}

func (c *ModCache) LookupReplace(importPath string) (matched modfile.PathIdentity, replace modfile.PathIdentity, exists bool) {
	for _, path := range paths(importPath) {
		p := modfile.PathIdentity{Path: path}
		if rp, ok := c.replace[p]; ok {
			return p, rp.PathIdentity, true
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

	// cached moduel@tag too
	if mod.TagVersion != "" {
		c.mods[mod.Module+"@"+mod.TagVersion] = mod
	}

	c.SetRepoVersion(mod.Repo, mod.ModVersion)

	for repo, r := range mod.Require {
		c.SetRepoVersion(repo, r.ModVersion)
	}

	for k, replaceTarget := range mod.Replace {
		if currentReplaceTarget, ok := c.replace[k]; !ok {
			c.replace[k] = pathIdentityWithMod{mod: mod, PathIdentity: replaceTarget}
		} else {
			if replaceTarget.String() != currentReplaceTarget.PathIdentity.String() {
				fmt.Printf(`
[WARNING] '%s' already replaced to 
	'%s' (using by module '%s'), but another module want to replace as 
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

func (c *ModCache) SetRepoVersion(module string, version modfile.ModVersion) {
	if mv, ok := c.repoVersions[module]; ok {
		if mv.Version == "" {
			c.repoVersions[module] = version
		} else if versionGreaterThan(version.Version, mv.Version) {
			c.repoVersions[module] = version
		} else if version.Version == mv.Version && version.TagVersion != "" {
			// sync tag version
			mv.TagVersion = version.TagVersion
			c.repoVersions[module] = mv
		}

	} else {
		c.repoVersions[module] = version
	}
}

func (c *ModCache) RepoVersion(repo string) modfile.ModVersion {
	if v, ok := c.repoVersions[repo]; ok {
		return v
	}
	return modfile.ModVersion{}
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

	return c.get(ctx, repo, version, importPath)
}

const versionUpgrade = "upgrade"

func (c *ModCache) get(ctx context.Context, repo string, requestedVersion string, importPath string) (*Mod, error) {
	version := requestedVersion

	if version == "" {
		version = versionUpgrade
	}

	if OptsFromContext(ctx).Upgrade {
		version = versionUpgrade

		// when tag version exists, should upgrade with tag version
		if mv, ok := c.repoVersions[repo]; ok {
			if mv.TagVersion != "" {
				version = mv.TagVersion
			}
		}
	} else {
		// use the resolved version, when already resolved.
		if mv, ok := c.repoVersions[repo]; ok {
			if mv.TagVersion != "" && mv.Version != "" && mv.TagVersion == requestedVersion {
				version = mv.Version
			}
		}
	}

	// mod@version replace
	if r, ok := c.replace[modfile.PathIdentity{Path: repo, Version: version}]; ok {
		repo, version = r.Path, r.Version
	}

	if version == "" {
		version = versionUpgrade
	}

	var root *Mod

	if mod, ok := c.mods[repo+"@"+version]; ok && mod.Resolved() {
		root = mod
	} else {
		logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("get %s@%s", repo, version))

		m, err := c.download(repo, version)
		if err != nil {
			return nil, err
		}

		if version != versionUpgrade {
			m.TagVersion = requestedVersion
		}

		root = m

		if _, err := root.LoadInfo(); err != nil {
			return nil, err
		}

		c.Collect(ctx, root)
	}

	if root != nil {

		// sub dir may as mod.
		importPaths := paths(importPath)

		for _, module := range importPaths {
			if module == root.Module {
				break
			}

			if mod, ok := c.mods[module+"@"+version]; ok && mod.Resolved() {
				return mod, nil
			} else {
				rel, _ := subPath(root.Module, module)

				sub := Mod{}
				sub.Repo = root.Repo
				sub.RepoSum = root.RepoSum

				sub.Module = module
				sub.ModVersion = root.ModVersion

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

	c.SetRepoVersion(r.Root, modfile.ModVersion{})

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
	mod.RepoSum = info.Sum

	return mod, nil
}
