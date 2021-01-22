package glob

import (
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v2"
)

type OptFunc func(o *opt)

type opt struct {
	Dir string
}

func Dir(dir string) OptFunc {
	return func(o *opt) {
		o.Dir = dir
	}
}

func Src(patterns []string, optFns ...OptFunc) ([]string, error) {
	o := opt{}

	for i := range optFns {
		optFns[i](&o)
	}

	fileSet := map[string]bool{}

	for i := range patterns {
		pattern := patterns[i]
		omit := pattern[0] == '!'
		if omit {
			pattern = pattern[1:]
		}

		p := pattern

		if o.Dir != "" {
			p = filepath.Join(o.Dir, pattern)
		}

		matched, err := doublestar.Glob(p)
		if err != nil {
			return nil, err
		}
		for _, f := range matched {
			fileSet[f] = !omit
		}
	}

	files := make([]string, 0)

	for f, matched := range fileSet {
		if matched {
			files = append(files, f)
		}
	}

	sort.Strings(files)

	return files, nil
}
