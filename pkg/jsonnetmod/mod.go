package jsonnetmod

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/go-jsonnet/ast"

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

	if m.Comments == nil {
		m.Comments = map[string][]string{}
	}

	if m.Replace == nil {
		m.Replace = map[PathIdentity]PathIdentity{}
	}

	if m.Require == nil {
		m.Require = map[string]Require{}
	}

	if data, err := ioutil.ReadFile(f); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		node, err := jsonnet.SnippetToAST(f, string(data))
		if err != nil {
			return err
		}

		if o, ok := node.(*ast.DesugaredObject); ok {
			for _, f := range o.Fields {
				fieldName := f.Name.(*ast.LiteralString).Value

				switch fieldName {
				case "module":
					if v, ok := f.Body.(*ast.LiteralString); ok {
						m.Module = v.Value
					} else {
						return fmt.Errorf("%s must be a string value of %s", fieldName, modJsonnetFile)
					}
				case "jpath":
					if v, ok := f.Body.(*ast.LiteralString); ok {
						m.JPath = v.Value
					} else {
						return fmt.Errorf("%s must be a string value of %s", fieldName, modJsonnetFile)
					}
				case "replace":
					if v, ok := f.Body.(*ast.DesugaredObject); ok {
						if err := rangeField(v, func(name string, value string, hidden bool, f ast.DesugaredObjectField) error {
							from, err := ParsePathIdentity(name)
							if err != nil {
								return nil
							}

							to, err := ParsePathIdentity(value)
							if err != nil {
								return nil
							}

							if to.Path == "" {
								to.Path = from.Path
							}

							m.Replace[*from] = *to
							m.Comments["replace:"+name] = pickNodeComponents(f.Name)

							return nil
						}); err != nil {
							return err
						}
					} else {
						return fmt.Errorf("%s must be a object of %s", fieldName, modJsonnetFile)
					}
				case "require":
					if v, ok := f.Body.(*ast.DesugaredObject); ok {
						if err := rangeField(v, func(name string, value string, hidden bool, f ast.DesugaredObjectField) error {
							m.Require[name] = Require{
								Version:  value,
								Indirect: hidden,
							}
							m.Comments["require:"+name] = pickNodeComponents(f.Name)
							return nil
						}); err != nil {
							return err
						}
					} else {
						return fmt.Errorf("%s must be a object of %s", fieldName, modJsonnetFile)
					}
				}
			}
		} else {
			return fmt.Errorf("invalid %s", modJsonnetFile)
		}

		return nil
	}

	return nil
}

func pickNodeComponents(node ast.Node) []string {
	comments := make([]string, 0)

	if f := node.OpenFodder(); f != nil {
		for _, fe := range *f {
			comments = append(comments, fe.Comment...)
		}
	}

	return comments
}

func rangeField(o *ast.DesugaredObject, each func(name string, value string, hidden bool, f ast.DesugaredObjectField) error) error {
	for _, f := range o.Fields {
		key, ok := f.Name.(*ast.LiteralString)
		if !ok {
			return fmt.Errorf("%s should be a string", f.Name)
		}

		value, ok := f.Body.(*ast.LiteralString)
		if !ok {
			return fmt.Errorf("%s should be a string", f.Body)
		}
		if err := each(key.Value, value.Value, f.Hide == ast.ObjectFieldHidden, f); err != nil {
			return err
		}
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

			writeBlock(buf, func() {
				replaces := make([]string, 0)
				for r := range m.Replace {
					replaces = append(replaces, r.String())
				}
				sort.Strings(replaces)

				for _, f := range replaces {
					from, _ := ParsePathIdentity(f)

					if comments, ok := m.Comments["replace:"+from.String()]; ok {
						for _, c := range comments {
							_, _ = fmt.Fprintln(buf, c)
						}
					}

					to := m.Replace[*from]

					if to.Path == from.Path {
						to.Path = ""
					}

					_, _ = fmt.Fprintf(buf, "'%s': '%s',\n", from, to)
				}
			}, false)
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

					if comments, ok := m.Comments["require:"+pkg]; ok {
						for _, c := range comments {
							_, _ = fmt.Fprintln(buf, c)
						}
					}

					if r.Indirect {
						_, _ = fmt.Fprintf(buf, "'%s':: '%s',\n", pkg, r.Version)
					} else {
						_, _ = fmt.Fprintf(buf, "'%s': '%s',\n", pkg, r.Version)
					}
				}
			}, false)
		}
	}, true)

	d, err := formatter.Format(f, buf.String(), formatter.DefaultOptions())
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f, []byte(d), os.ModePerm)
}

func writeBlock(buf io.Writer, next func(), root bool) {
	_, _ = fmt.Fprintln(buf, "{")
	next()
	_, _ = fmt.Fprintf(buf, "}")
	if !root {
		_, _ = fmt.Fprint(buf, ",")
	}
	_, _ = fmt.Fprintln(buf, "")
}

type Mod struct {
	// Module name
	Module string
	// JPath JSONNET_PATH
	// when not empty, symlinks will be created for JSONNET_PATH
	JPath string
	// Replace
	// version limit
	Replace map[PathIdentity]PathIdentity
	// Require same as go root
	// require { module: version }
	// indirect require { module:: version }
	Require map[string]Require

	// Version
	Version string
	// Dir
	Dir string
	// Sum
	Sum string

	Comments map[string][]string
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
