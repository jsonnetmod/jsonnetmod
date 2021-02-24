package forkinternal

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/octohelm/jsonnetmod/pkg/util"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
)

func ForkInternals(dir string, importPaths ...string) error {
	task, err := TaskFor(dir)
	if err != nil {
		return err
	}

	for _, importPath := range importPaths {
		if err := task.ForkInternalFromImportPath(importPath); err != nil {
			return err
		}
	}

	return nil
}

func TaskFor(dir string) (*Task, error) {
	if !filepath.IsAbs(dir) {
		cwd, _ := os.Getwd()
		dir = path.Join(cwd, dir)
	}

	d := dir

	for d != "/" {
		gmodfile := filepath.Join(d, "go.mod")

		if data, err := os.ReadFile(gmodfile); err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
		} else {
			f, _ := modfile.Parse(gmodfile, data, nil)

			rel, _ := filepath.Rel(d, dir)
			return &Task{
				PkgPath:     filepath.Join(f.Module.Mod.Path, rel),
				Dir:         filepath.Join(d, rel),
				Forked:      map[string]bool{},
				ForkedFiles: map[string]string{},
			}, nil
		}

		d = filepath.Join(d, "../")
	}

	return nil, fmt.Errorf("missing go.mod")
}

type Task struct {
	Dir         string
	PkgPath     string
	Forked      map[string]bool
	ForkedFiles map[string]string
}

func replaceInternal(p string) string {
	return strings.Replace(p, "internal/", "internalpkg/", -1)
}

func (t *Task) ForkInternalFromImportPath(importPath string) error {
	if _, ok := t.Forked[importPath]; ok {
		return nil
	}
	defer func() {
		t.Forked[importPath] = true
	}()

	pkg, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return err
	}

	if err := t.ForkInternal(pkg); err != nil {
		return err
	}

	return nil
}

func (t *Task) ForkInternal(pkg *build.Package) error {
	files, err := filepath.Glob(pkg.Dir + "/*.go")
	if err != nil {
		return err
	}

	outputs := make(map[string]bool)

	outputDir := ""

	for _, f := range files {
		// skip test file
		if strings.HasSuffix(f, "_test.go") {
			continue
		}

		output, err := t.forkGoFile(f, pkg)
		if err != nil {
			return err
		}

		if output == "" {
			continue
		}

		outputs[output] = true

		outputDir = filepath.Dir(output)
	}

	// cleanup
	if outputDir != "" {
		files, err := filepath.Glob(outputDir + "/*.go")
		if err != nil {
			return err
		}

		for _, f := range files {
			if _, ok := outputs[f]; !ok {
				if err := os.RemoveAll(f); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *Task) forkGoFile(filename string, pkg *build.Package) (string, error) {
	if o, ok := t.ForkedFiles[filename]; ok {
		return o, nil
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return "", err
	}

	if pkg.Name == "" {
		if file.Name.Name != "main" {
			pkg.Name = file.Name.Name
		}
	}

	if file.Name.Name != pkg.Name {
		return "", nil
	}

	for _, i := range file.Imports {
		importPath, _ := strconv.Unquote(i.Path.Value)

		if strings.Contains(importPath, "internal/") {
			if err := t.ForkInternalFromImportPath(importPath); err != nil {
				return "", err
			}
			_ = astutil.RewriteImport(fset, file, importPath, filepath.Join(t.PkgPath, replaceInternal(importPath)))
		}
	}

	buf := bytes.NewBuffer(nil)

	if err := format.Node(buf, fset, file); err != nil {
		return "", err
	}

	output := filepath.Join(t.Dir, replaceInternal(pkg.ImportPath), filepath.Base(filename))

	t.ForkedFiles[filename] = output

	return output, util.WriteFile(output, buf.Bytes())
}
