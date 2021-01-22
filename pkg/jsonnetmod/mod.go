package jsonnetmod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/formatter"
)

const modJsonnetFile = "mod.jsonnet"

func ModFromDir(dir string) (*Mod, error) {
	m := &Mod{}
	m.Dir = dir
	m.Version = "latest"

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "%s not found", dir)
	}

	if err := LoadModInfo(m); err != nil {
		return nil, err
	}

	return m, nil
}

func LoadModInfo(m *Mod) error {
	f := filepath.Join(m.Dir, modJsonnetFile)

	if m.Replace == nil {
		m.Replace = map[PathReplace]PathReplace{}
	}

	if m.Require == nil {
		m.Require = map[string]Require{}
	}

	if data, err := ioutil.ReadFile(f); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		vm := jsonnet.MakeVM()

		jsonRaw, err := vm.EvaluateAnonymousSnippet(f, fmt.Sprintf(`
local d = (%s);
local convertRequire = function (o = {}) { [pkg]: { Version: o[pkg], Indirect: !std.objectHas(o, pkg) } for pkg in std.objectFieldsAll(o) } ;

( if std.objectHas(d, 'module') then { Module: d.module } else {} )
 + ( if std.objectHas(d, 'jpath') then { JPath: d.jpath } else {} )
 + ( if std.objectHas(d, 'replace') then { Replace: d.replace } else {} )
 + ( if std.objectHas(d, 'require') then { Require: convertRequire(d.require) } else {} )
`, data))
		if err != nil {
			return err
		}

		return json.Unmarshal([]byte(jsonRaw), m)
	}

	return nil
}

func WriteMod(m *Mod) error {
	f := filepath.Join(m.Dir, modJsonnetFile)

	buf := bytes.NewBuffer(nil)

	writeBlock(buf, func() {
		_, _ = fmt.Fprintf(buf, "module: '%s',\n", m.Module)

		if m.JPath != "" {
			_, _ = fmt.Fprintf(buf, "jpath: '%s',\n", m.JPath)
		}

		if len(m.Replace) > 0 {
			_, _ = fmt.Fprintf(buf, "replace: ")

			data, _ := json.MarshalIndent(m.Replace, "", "  ")

			_, _ = fmt.Fprintf(buf, "%s,\n", data)
		}

		if len(m.Require) > 0 {
			_, _ = fmt.Fprintf(buf, "require: ")

			writeBlock(buf, func() {
				pkgs := make([]string, 0)
				for pkg := range m.Require {
					pkgs = append(pkgs, pkg)
				}
				sort.Strings(pkgs)

				for _, pkg := range pkgs {
					r := m.Require[pkg]
					if r.Indirect {
						_, _ = fmt.Fprintf(buf, "'%s':: '%s',\n", pkg, r.Version)
					} else {
						_, _ = fmt.Fprintf(buf, "'%s': '%s',\n", pkg, r.Version)
					}
				}
			})

			_, _ = fmt.Fprintf(buf, ",\n")
		}
	})

	d, err := formatter.Format(f, buf.String(), formatter.DefaultOptions())
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f, []byte(d), os.ModePerm)
}

func writeBlock(buf io.Writer, next func()) {
	_, _ = fmt.Fprintln(buf, "{")
	next()
	_, _ = fmt.Fprintln(buf, "}")
}

type Mod struct {
	// Module name
	Module string `json:",omitempty"`
	// JPath JSONNET_PATH
	// when not empty, symlinks will be created for JSONNET_PATH
	JPath string `json:",omitempty"`
	// Replace
	// version limit
	Replace map[PathReplace]PathReplace `json:",omitempty"`
	// Require same as go root
	// require { module: version }
	// indirect require { module:: version }
	Require map[string]Require `json:",omitempty"`

	// Version
	Version string `json:"-"`
	// Dir
	Dir string `json:"-"`
	// Sum
	Sum string `json:"-"`
}

func (m *Mod) Clone() *Mod {
	mod := &Mod{}
	mod.Dir = m.Dir
	mod.Version = m.Version
	mod.Sum = m.Sum
	mod.Module = m.Module
	mod.JPath = m.JPath

	mod.Replace = map[PathReplace]PathReplace{}
	for k, v := range m.Replace {
		mod.Replace[k] = v
	}

	mod.Require = map[string]Require{}
	for k, v := range m.Require {
		mod.Require[k] = v
	}

	return mod
}

type Require struct {
	Version  string
	Indirect bool `json:",omitempty"`
}

func (m *Mod) Resolved() bool {
	return m.Dir != ""
}

func (m *Mod) SetRequire(module string, version string, indirect bool) {
	if module == m.Module {
		return
	}

	if m.Require == nil {
		m.Require = map[string]Require{}
	}

	r := Require{Version: version, Indirect: indirect}

	if currentRequire, ok := m.Require[module]; ok {
		// always using greater one
		if versionGreaterThan(currentRequire.Version, r.Version) {
			r.Version = currentRequire.Version
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

	if matched, replace, ok := cache.LookupReplace(importPath, version); ok {
		// xxx => ../xxx
		if replace.IsLocalReplace() {
			mod, err := ModFromDir(filepath.Join(m.Dir, replace.Path))
			if err != nil {
				return nil, err
			}
			cache.Set(mod)
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
