package vmod

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"golang.org/x/mod/modfile"
)

type Mod struct {
	Requirements map[string]string `yaml:"requirements,omitempty"`
	Aliases      map[string]string `yaml:"aliases,omitempty"`
}

func (m *Mod) SyncFromGoMod(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	gomod, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return err
	}

	m.Requirements = map[string]string{}

	for i := range gomod.Require {
		r := gomod.Require[i]
		m.Requirements[r.Mod.Path] = r.Mod.Version
	}

	return nil
}

func (m *Mod) ToGoMod() []byte {
	b := bytes.NewBuffer(nil)

	_, _ = fmt.Fprintln(b, "module vendor")

	if len(m.Requirements) > 0 {
		_, _ = fmt.Fprintln(b, "require (")

		for r, version := range m.Requirements {
			_, _ = fmt.Fprintf(b, "\t%s %s\n", r, version)
		}

		_, _ = fmt.Fprintln(b, ")")
	}

	return b.Bytes()
}
