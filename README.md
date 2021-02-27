# Jsonnet Mod

[![GoDoc Widget](https://godoc.org/github.com/jsonnetmod/jsonnetmod?status.svg)](https://godoc.org/github.com/jsonnetmod/jsonnetmod)
[![codecov](https://codecov.io/gh/jsonnetmod/jsonnetmod/branch/master/graph/badge.svg)](https://codecov.io/gh/jsonnetmod/jsonnetmod)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsonnetmod/jsonnetmod)](https://goreportcard.com/report/github.com/jsonnetmod/jsonnetmod)

[JSONNET](https://jsonnet.org/) dependency management based on [go modules](https://golang.org/ref/mod)

## Requirements

* `git` or other vcs tool supported by go for vcs downloading.

## Install

```shell
go install github.com/jsonnetmod/jsonnetmod/cmd/jmod@latest
```

## Usage

```shell 
# build, will automately install deps if not exists. 
jmod build ./path/to/jsonnetfile

# auto dep
jmod get ./...

# upgrade dep
jmod get -u ./...

# install dep with special version
jmod get github.com/grafana/jsonnet-libs@latest
```

## Features

* Dependency management based on go modules
    * all dependency codes will download under `$(go env GOMODCACHE)`
    * `GOPROXY` supported to speed up downloading
    * `JSONNET_PATH` compatibility
* Compatible with `jsonnetfile.json`
    * not support local source, use `replace` to fix 
* Includes all [native functions of tanka](https://tanka.dev/jsonnet/native)
* Object YAML import supported
    * with hidden fields `__dirname` and `__filename`

## Spec `mod.jsonnet`

```jsonnet
{
  // module name
  // for sub mod import, <module>/path/to/local/file.jsonnet
  module: 'github.com/x/b',
  
  // `JSONNET_PATH` compatibility, 
  // when this field defined, all dependencies will created symlinks under ./<jpath>
  jpath: 'vendor',
  
  // dependency version lock or replace
  // only support dir
  replace: {
    // version lock
    // if path equals, could short as '@v0.0.18'
    'github.com/rancher/local-path-provisioner': 'github.com/rancher/local-path-provisioner@v0.0.18',
    // local mod replace
    'github.com/x/a': '../a',
    // import stubs
    'k.libsonnet': 'github.com/jsonnet-libs/k8s-alpha/1.19/main.libsonnet',
    // mod short alias
    'ksonnet-util': 'github.com/grafana/jsonnet-libs/ksonnet-util',
  },
  
  // automately resolve by the jsonnet code `import` or `importstr`
  // rules follow go modules
  // :: hidden fields means indirect require
  require: {
    'github.com/rancher/local-path-provisioner':: 'v0.0.19',
    // ,<tag_version>, when upgrade, should use tag version for upgrade. 
    'github.com/grafana/jsonnet-libs':: 'v0.0.0-20210209092858-49e80898b183,master',
    'github.com/x/a': 'v0.0.0',
  },
}
```

### Known issues

#### dep incompatible go mod repo

For some go project like 

```
$ go mod download -json github.com/grafana/loki@v2.1.0
{
        "Path": "github.com/grafana/loki",
        "Version": "v2.1.0",
        "Error": "github.com/grafana/loki@v2.1.0: invalid version: module contains a go.mod file, so major version must be compatible: should be v0 or v1, not v2"
}
```

Could config `mod.jsonnet` replace with commit hash of the tag to hack

```jsonnet
{
    replace: {
        'github.com/grafana/loki': '@1b79df3',
    }
}
```

## Plugin kube

Like Tanka but without struct limit.
Only one requirement, make sure the file return [`tanka.dev/Environment` object](https://tanka.dev/inline-environments#converting-to-an-inline-environment)

```
cd ./examples
jmod k show ./clusters/demo/hello-world.jsonnet
jmod k apply ./clusters/demo/hello-world.jsonnet
jmod k delete ./clusters/demo/hello-world.jsonnet
jmod k prune ./clusters/demo/hello-world.jsonnet
```

### Work with Tanka

```
jmod build -o ./environments/demo/main.jsonnet ./path/to/tanka-environment.jsonnet
tk show ./environments/demo
```

export not support pipe now, so need to create inline env object main.jsonnet first.
