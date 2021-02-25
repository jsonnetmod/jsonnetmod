package jsonnetmod

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jsonnetmod/jsonnetmod/pkg/jsonnetmod/jsonnetfile"
	"github.com/pkg/errors"

	"github.com/jsonnetmod/jsonnetmod/pkg/jsonnetmod/modfile"
)

type Mod struct {
	modfile.ModFile
	modfile.ModVersion
	// Repo
	Repo string
	// RepoSum
	RepoSum string
	// Dir
	Dir string
}

func (m *Mod) String() string {
	if m.Version == "" {
		return m.Module + "@v0.0.0"
	}
	return m.Module + "@" + m.Version
}

func (m *Mod) LoadInfo() (bool, error) {
	if _, err := os.Stat(m.Dir); os.IsNotExist(err) {
		return false, errors.Wrapf(err, "%s not found", m.Dir)
	}

	// mod.jsonnet
	modfileExists, err := modfile.LoadModFile(m.Dir, &m.ModFile)
	if err != nil {
		return false, err
	}

	if !modfileExists {
		// jsonnetfile.json
		jsonnetfileExists, err := jsonnetfile.LoadModFile(m.Dir, &m.ModFile)
		if err != nil {
			return false, err
		}

		if jsonnetfileExists {
			return jsonnetfileExists, nil
		}
	}

	return modfileExists, nil
}

func (m *Mod) Resolved() bool {
	return m.Dir != ""
}

func (m *Mod) SetRequire(module string, modVersion modfile.ModVersion, indirect bool) {
	if module == m.Module {
		return
	}

	if m.Require == nil {
		m.Require = map[string]modfile.Require{}
	}

	r := modfile.Require{}
	r.ModVersion = modVersion
	r.Indirect = indirect

	if currentRequire, ok := m.Require[module]; ok {
		// always using greater one
		if versionGreaterThan(currentRequire.Version, r.Version) {
			r.ModVersion = currentRequire.ModVersion
		}

		if r.Indirect {
			r.Indirect = currentRequire.Indirect
		}
	}

	m.Require[module] = r
}

func (m *Mod) ResolveImportPath(ctx context.Context, cache *ModCache, importPath string, version string) (*ImportPath, error) {
	// self import '<mod.module>/dir/to/sub'
	if isSubDirFor(importPath, m.Module) {
		return ImportPathFor(m, importPath), nil
	}

	if matched, replace, ok := cache.LookupReplace(importPath); ok {
		// xxx => ../xxx
		if replace.IsLocalReplace() {
			mod := &Mod{Dir: filepath.Join(m.Dir, replace.Path)}
			mod.Version = "v0.0.0"
			if _, err := mod.LoadInfo(); err != nil {
				return nil, err
			}
			cache.Collect(ctx, mod)
			return ImportPathFor(mod, importPath), nil
		}

		// a[@latest] => b@latest
		// must strict version
		replacedImportPath := replaceImportPath(replace.Path, matched.Path, importPath)

		ctxWithUpgradeDisabled := WithOpts(ctx, OptUpgrade(false))

		fixVersion := m.fixVersion

		if replace.Version != "" {
			fixVersion = nil
		}

		mod, err := cache.Get(ctxWithUpgradeDisabled, replacedImportPath, replace.Version, fixVersion)
		if err != nil {
			return nil, err
		}

		return ImportPathFor(mod, replacedImportPath), nil
	}

	mod, err := cache.Get(ctx, importPath, version, m.fixVersion)
	if err != nil {
		return nil, err
	}

	return ImportPathFor(mod, importPath), nil
}

func (m *Mod) fixVersion(repo string, version string) string {
	if m.Require != nil {
		if r, ok := m.Require[repo]; ok {
			return r.Version
		}
	}
	return version
}
