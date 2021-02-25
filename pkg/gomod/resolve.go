package gomod

import (
	"context"

	"github.com/jsonnetmod/jsonnetmod/internal/forked/cmd/go/internalpkg/cfg"
	"github.com/jsonnetmod/jsonnetmod/internal/forked/cmd/go/internalpkg/modfetch"
	"github.com/jsonnetmod/jsonnetmod/internal/forked/cmd/go/internalpkg/modload"
	"golang.org/x/mod/module"
)

type Module struct {
	Path     string
	Version  string
	Error    string
	Info     string
	GoMod    string
	Zip      string
	Dir      string
	Sum      string
	GoModSum string
}

func ResolveModule(ctx context.Context, path string, version string, verbose bool) *Module {
	modload.ForceUseModules = true
	cfg.BuildX = verbose

	found := modload.ListModules(ctx, []string{path + "@" + version}, false, false, false)

	downloadModule := func(m *Module) {
		var err error
		m.Info, err = modfetch.InfoFile(m.Path, m.Version)
		if err != nil {
			m.Error = err.Error()
			return
		}
		m.GoMod, err = modfetch.GoModFile(m.Path, m.Version)
		if err != nil {
			m.Error = err.Error()
			return
		}
		m.GoModSum, err = modfetch.GoModSum(m.Path, m.Version)
		if err != nil {
			m.Error = err.Error()
			return
		}
		mod := module.Version{Path: m.Path, Version: m.Version}
		m.Zip, err = modfetch.DownloadZip(ctx, mod)
		if err != nil {
			m.Error = err.Error()
			return
		}
		m.Sum = modfetch.Sum(mod)
		m.Dir, err = modfetch.Download(ctx, mod)
		if err != nil {
			m.Error = err.Error()
			return
		}
	}

	if len(found) > 0 {
		info := found[0]

		m := &Module{
			Path:    info.Path,
			Version: info.Version,
		}

		if info.Error != nil {
			m.Error = info.Error.Err
		} else {
			downloadModule(m)
		}

		return m
	}

	return nil
}
