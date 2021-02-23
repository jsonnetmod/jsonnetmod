package modfile

import (
	"fmt"
	"strings"
)

const ModFilename = "mod.jsonnet"

type ModFile struct {
	// Module name
	Module string
	// JPath JSONNET_PATH
	// when not empty, symlinks will be created for JSONNET_PATH
	JPath string
	// Replace
	// version limit
	Replace map[PathIdentity]PathIdentity
	// Require same as go root
	// require { module: version }
	// indirect require { module:: version }
	Require map[string]Require
	// Comments
	Comments map[string][]string
}

// v0.0.0,v
func ParseModVersion(v string) ModVersion {
	mv := ModVersion{}

	versions := strings.Split(v, ",")

	mv.Version = versions[0]

	if len(versions) > 1 {
		mv.TagVersion = versions[1]
	} else {
		mv.TagVersion = mv.Version
	}

	return mv
}

type ModVersion struct {
	Version    string
	TagVersion string
}

func (mv ModVersion) String() string {
	if mv.TagVersion == "" || mv.Version == mv.TagVersion {
		return mv.Version
	}
	return mv.Version + "," + mv.TagVersion
}

type Require struct {
	ModVersion
	Indirect bool `json:",omitempty"`
}

func ParsePathIdentity(v string) (*PathIdentity, error) {
	if len(v) == 0 {
		return nil, fmt.Errorf("invalid %s", v)
	}

	parts := strings.Split(v, "@")

	i := parts[0]

	if i != "" && i[0] == '.' {
		return &PathIdentity{Path: i}, nil
	}

	if len(parts) > 1 {
		return &PathIdentity{Path: i, Version: parts[1]}, nil
	}
	return &PathIdentity{Path: i}, nil

}

type PathIdentity struct {
	Version string
	Path    string
}

func (r *PathIdentity) UnmarshalText(text []byte) error {
	rp, err := ParsePathIdentity(string(text))
	if err != nil {
		return err
	}
	*r = *rp
	return nil
}

func (r PathIdentity) MarshalText() (text []byte, err error) {
	return []byte(r.String()), nil
}

func (r PathIdentity) IsLocalReplace() bool {
	return len(r.Path) > 0 && r.Path[0] == '.'
}

func (r PathIdentity) String() string {
	if r.IsLocalReplace() {
		return r.Path
	}
	if r.Version != "" {
		return r.Path + "@" + r.Version
	}
	return r.Path
}
