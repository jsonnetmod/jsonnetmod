package jsonnetfile

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1"
	"github.com/octohelm/jsonnetmod/pkg/jsonnetmod/modfile"
)

func LoadModFile(dir string, m *modfile.ModFile) (bool, error) {
	f := filepath.Join(dir, "jsonnetfile.json")

	if m.Comments == nil {
		m.Comments = map[string][]string{}
	}

	if m.Replace == nil {
		m.Replace = map[modfile.PathIdentity]modfile.PathIdentity{}
	}

	if data, err := ioutil.ReadFile(f); err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
	} else {
		jf := &spec.JsonnetFile{}

		if err := jf.UnmarshalJSON(data); err != nil {
			return false, err
		}

		for _, d := range jf.Dependencies {
			if d.Source.GitSource != nil {
				repo := filepath.Join(d.Source.GitSource.Host, d.Source.GitSource.User, d.Source.GitSource.Repo)

				m.Replace[modfile.PathIdentity{Path: repo}] = modfile.PathIdentity{Path: repo, Version: d.Version}

				if jf.LegacyImports {
					m.Replace[modfile.PathIdentity{Path: d.LegacyName()}] = modfile.PathIdentity{Path: filepath.Join(repo, d.Source.GitSource.Subdir)}
				}
			}
		}

		return true, nil
	}

	return false, nil
}
