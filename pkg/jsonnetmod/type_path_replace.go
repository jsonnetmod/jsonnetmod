package jsonnetmod

import (
	"fmt"
	"strings"
)

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
