package jsonnetmod

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/google/go-jsonnet"
)

type PathResolver interface {
	Resolve(ctx context.Context, importPath string, importedFrom string) (absDir string, err error)
}

func NewImporter(ctx context.Context, resolver PathResolver) jsonnet.Importer {
	return &Importer{ctx: ctx, pathResolver: resolver}
}

type Importer struct {
	ctx          context.Context
	fi           jsonnet.FileImporter
	pathResolver PathResolver
}

func (i *Importer) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	// todo hack the `k.libsonnet`
	if importedPath == "k.libsonnet" {
		importedPath = "k/main.libsonnet"
	}

	abs := importedPath

	if !filepath.IsAbs(importedPath) {
		abs = filepath.Join(filepath.Dir(importedFrom), importedPath)
	}

	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) && importedPath[0] == '.' {
			return jsonnet.Contents{}, "", err
		}

		p := filepath.Dir(importedPath)
		filename := filepath.Base(importedPath)

		dir, err := i.pathResolver.Resolve(i.ctx, p, importedFrom)
		if err != nil {
			return jsonnet.Contents{}, "", errors.Wrapf(err, "resolve failed `%s`", importedPath)
		}

		abs = filepath.Join(dir, filename)
	}

	ext := filepath.Ext(abs)

	// todo refactor as loader like webpack
	switch ext {
	case ".yaml", ".yml":
		var v map[string]interface{}
		if err := readYAMLFile(abs, &v); err != nil {
			return jsonnet.Contents{}, "", err
		}
		data, _ := json.MarshalIndent(v, "", "  ")
		return jsonnet.MakeContents(fmt.Sprintf("%s + { __dirname:: '%s' }", data, filepath.Dir(abs))), abs, nil
	case ".json", ".jsonnet", ".libsonnet":
		return i.fi.Import(importedFrom, abs)
	}
	return jsonnet.Contents{}, "", fmt.Errorf("unsupport modfile ext `%s`", ext)
}
