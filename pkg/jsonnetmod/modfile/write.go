package modfile

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/go-jsonnet/formatter"
)

func WriteModFile(dir string, m *ModFile) error {
	f := filepath.Join(dir, ModFilename)

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
						_, _ = fmt.Fprintf(buf, "'%s':: '%s',\n", pkg, r.ModVersion)
					} else {
						_, _ = fmt.Fprintf(buf, "'%s': '%s',\n", pkg, r.ModVersion)
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
