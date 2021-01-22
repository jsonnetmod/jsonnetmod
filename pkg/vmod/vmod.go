package vmod

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/mod/module"
	"golang.org/x/tools/go/vcs"
)

func VModFor(root string) (*VMod, error) {
	vm := &VMod{}

	vm.root = root
	vm.vendorRoot = filepath.Join(root, "vendor")

	vm.modFile = filepath.Join(root, "vmod.yaml")

	vm.gomod.file = filepath.Join(vm.vendorRoot, "go.mod")
	if cache, err := resolveGomodCache(); err != nil {
		return nil, err
	} else {
		vm.gomod.cache = cache
	}

	if err := os.MkdirAll(vm.vendorRoot, os.ModePerm); err != nil {
		return nil, err
	}

	if err := readYAMLFile(vm.modFile, &vm.mod); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return vm, nil
}

type VMod struct {
	root string

	modFile    string
	vendorRoot string

	gomod struct {
		cache string
		file  string
	}

	mod Mod
}

func (v *VMod) ensureGomod() error {
	return ioutil.WriteFile(v.gomod.file, v.mod.ToGoMod(), os.ModePerm)
}

func (v *VMod) postInstall(ctx context.Context) error {
	if err := v.mod.SyncFromGoMod(v.gomod.file); err != nil {
		return err
	}
	if err := v.ensureSymlinks(ctx); err != nil {
		return err
	}
	return writeYAMLFile(v.modFile, v.mod)
}

func (v *VMod) modPath(importPath string, version string) (string, error) {
	i, err := module.EscapePath(importPath)
	if err != nil {
		return "", errors.Wrapf(err, "invalid import path %s", importPath)
	}
	ver, err := module.EscapeVersion(version)
	if err != nil {
		return "", errors.Wrapf(err, "invalid version %s of %s", version, importPath)
	}
	return filepath.Join(v.gomod.cache, i+"@"+ver), nil
}

func (v *VMod) ensureSymlinks(ctx context.Context) error {
	links := make([][2]string, 0)
	repoRoots := map[string]string{}

	for importPath, version := range v.mod.Requirements {
		modPath, err := v.modPath(importPath, version)
		if err != nil {
			return err
		}

		links = append(links, [2]string{
			modPath,
			filepath.Join(v.vendorRoot, importPath),
		})

		repoRoots[importPath] = modPath
	}

	for to, from := range v.mod.Aliases {
		for repo, repoRoot := range repoRoots {
			if strings.HasPrefix(from, repo) {
				links = append(links, [2]string{
					filepath.Join(repoRoot, "."+from[len(repo):]),
					filepath.Join(v.vendorRoot, to),
				})
			}
		}
	}

	matches, _ := filepath.Glob(v.vendorRoot + "/*")

	for _, p := range matches {
		f, err := os.Lstat(p)
		if err != nil {
			return err
		}

		if !f.IsDir() {
			continue
		}

		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}

	for _, link := range links {
		if err := symlink(link[0], link[1]); err != nil {
			return err
		}
	}

	return nil
}

func (v *VMod) Alias(ctx context.Context, to string, importPath string) error {
	if strings.Contains(to, string([]byte{filepath.Separator})) {
		return fmt.Errorf("invalid alias %s: only support flatten", to)
	}

	for repo := range v.mod.Requirements {
		if strings.HasPrefix(importPath, repo) {
			if v.mod.Aliases == nil {
				v.mod.Aliases = map[string]string{}
			}
			v.mod.Aliases[to] = importPath
			return v.postInstall(ctx)
		}
	}
	return fmt.Errorf("invalid import path %s", importPath)
}

func (v *VMod) Get(ctx context.Context, importPathMayWithVersion string) error {
	if err := v.ensureGomod(); err != nil {
		return err
	}

	importRepoPath, err := v.resolveRepoFromImportPath(ctx, importPathMayWithVersion)
	if err != nil {
		return err
	}

	if err := v.get(ctx, importRepoPath); err != nil {
		return err
	}
	return v.postInstall(ctx)
}

func (v *VMod) get(ctx context.Context, importRepoPath string) error {
	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("get %s", importRepoPath))
	return v.exec(ctx, fmt.Sprintf("get -v %s", importRepoPath))
}

func (v *VMod) Upgrade(ctx context.Context) error {
	if err := v.ensureGomod(); err != nil {
		return err
	}
	for repo := range v.mod.Requirements {
		if err := v.get(ctx, repo); err != nil {
			return err
		}
	}
	return v.postInstall(ctx)
}

func (v *VMod) Download(ctx context.Context) error {
	if err := v.ensureGomod(); err != nil {
		return err
	}
	if err := v.exec(ctx, "mod download -x"); err != nil {
		return err
	}
	return v.postInstall(ctx)
}

func (v *VMod) resolveRepoFromImportPath(ctx context.Context, importRepoPath string) (string, error) {
	parts := strings.Split(importRepoPath, "@")

	importRepoPath, err := v.importRepoPath(ctx, parts[0])
	if err != nil {
		return "", err
	}

	if len(parts) > 1 {
		return importRepoPath + "@" + parts[1], nil
	}

	return importRepoPath, nil
}

func (v *VMod) importRepoPath(ctx context.Context, importPath string) (string, error) {
	for repo := range v.mod.Requirements {
		if strings.HasPrefix(importPath, repo) {
			return repo, nil
		}
	}

	logr.FromContextOrDiscard(ctx).V(1).Info(fmt.Sprintf("resolving %s", importPath))

	r, err := vcs.RepoRootForImportPath(importPath, true)
	if err != nil {
		return "", err
	}

	return r.Root, nil
}

func (v *VMod) exec(ctx context.Context, cmdline string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return errors.Wrapf(err, "missing %s", "go")
	}

	args := strings.Fields(cmdline)

	cmd := exec.Command("go", args...)

	cmd.Dir = v.vendorRoot
	cmd.Env = envForDir(cmd.Dir)

	log := logr.FromContextOrDiscard(ctx)

	go func() {
		stdout, _ := cmd.StdoutPipe()
		defer stdout.Close()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.V(1).Info(scanner.Text())
		}
	}()

	go func() {
		stderr, _ := cmd.StderrPipe()
		defer stderr.Close()

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.V(1).Info(scanner.Text())
		}
	}()

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "execute failed: %s", strings.Join(cmd.Args, " "))
	}

	return nil
}

func envForDir(dir string) []string {
	env := os.Environ()
	return mergeEnvLists([]string{"PWD=" + dir}, env)
}

func mergeEnvLists(in, out []string) []string {
NextVar:
	for _, inkv := range in {
		k := strings.SplitAfterN(inkv, "=", 2)[0]
		for i, outkv := range out {
			if strings.HasPrefix(outkv, k) {
				out[i] = inkv
				continue NextVar
			}
		}
		out = append(out, inkv)
	}
	return out
}

func resolveGomodCache() (string, error) {
	if _, err := exec.LookPath("go"); err != nil {
		return "", errors.Wrapf(err, "missing %s", "go")
	}

	d, err := exec.Command("go", "env", "GOMODCACHE").CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(d)), nil
}
