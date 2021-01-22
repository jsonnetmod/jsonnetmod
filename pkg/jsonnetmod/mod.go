package jsonnetmod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/formatter"
)

type Mod struct {
	// Module name
	Module string `json:",omitempty"`
	// JPath JSONNET_PATH
	// when not empty, symlinks will be created for JSONNET_PATH
	JPath string `json:",omitempty"`
	// Replace
	// ersion limit
	Replace map[PathReplace]PathReplace `json:",omitempty"`
	// Require same as go mod
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

type Require struct {
	Version  string
	Indirect bool `json:",omitempty"`
}

const modJsonnetFile = "mod.jsonnet"

func (m *Mod) Load() error {
	f := filepath.Join(m.Dir, modJsonnetFile)

	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	vm := jsonnet.MakeVM()
	jsonraw, err := vm.EvaluateAnonymousSnippet(f, fmt.Sprintf(`
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

	return json.Unmarshal([]byte(jsonraw), m)
}

func (m *Mod) Write() error {
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

func (m Mod) Lookup(importPath string) (*ImportPath, bool) {
	for pkg, r := range m.Require {
		if pathPrefix(importPath, pkg) {
			i := NewImportPath(pkg, r.Version, importPath)
			return i, true
		}
	}
	return nil, false
}

func (m *Mod) SetRequire(pkg string, version string, indirect bool) {
	if pkg == m.Module {
		return
	}

	if m.Require == nil {
		m.Require = map[string]Require{}
	}

	replace := Require{Version: version, Indirect: indirect}

	if r, ok := m.Require[pkg]; ok {
		if v := semver.Max(version, r.Version); v != "" {
			replace.Version = v
		}

		if !indirect {
			replace.Indirect = false
		}
	}

	m.Require[pkg] = replace
}

func (m Mod) LookupReplace(importPath string, version string) (matched PathReplace, replace PathReplace, exists bool) {
	for p, rp := range m.Replace {
		if pathPrefix(importPath, p.Path) {
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

func ParsePathIdentity(v string) (*PathReplace, error) {
	if len(v) == 0 {
		return nil, fmt.Errorf("invalid %s", v)
	}

	parts := strings.Split(v, "@")

	i := parts[0]

	if i[0] == '.' {
		return &PathReplace{Path: i}, nil
	}

	if len(parts) > 1 {
		return &PathReplace{Path: i, Version: parts[1]}, nil
	}
	return &PathReplace{Path: i}, nil

}

type PathReplace struct {
	Version string
	Path    string
}

func (r *PathReplace) UnmarshalText(text []byte) error {
	rp, err := ParsePathIdentity(string(text))
	if err != nil {
		return err
	}
	*r = *rp
	return nil
}

func (r PathReplace) MarshalText() (text []byte, err error) {
	return []byte(r.String()), nil
}

func (r PathReplace) IsLocalReplace() bool {
	return len(r.Path) > 0 && r.Path[0] == '.'
}

func (r PathReplace) String() string {
	if r.IsLocalReplace() {
		return r.Path
	}
	if r.Version != "" {
		return r.Path + "@" + r.Version
	}
	return r.Path
}

func NewImportPath(pkg string, version string, importPath string) *ImportPath {
	i := &ImportPath{}
	i.Module = pkg

	if version == "" {
		version = "latest"
	}

	i.Version = version

	if importPath != "" {
		i.SubPath, _ = subPath(i.Module, importPath)
	}

	return i
}

type ImportPath struct {
	Mod
	SubPath string
}

func (i ImportPath) Resolved() bool {
	return i.Dir != ""
}

func (i ImportPath) WithDir(dir string) *ImportPath {
	i.Dir = dir
	return &i
}

func (i *ImportPath) AbsolutePath() string {
	return filepath.Join(i.Dir, i.SubPath)
}
