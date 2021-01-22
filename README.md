# Jsonnet Mod

[![GoDoc Widget](https://godoc.org/github.com/octohelm/jsonnetmod?status.svg)](https://godoc.org/github.com/octohelm/jsonnetmod)
[![codecov](https://codecov.io/gh/octohelm/jsonnetmod/branch/master/graph/badge.svg)](https://codecov.io/gh/octohelm/jsonnetmod)
[![Go Report Card](https://goreportcard.com/badge/github.com/octohelm/jsonnetmod)](https://goreportcard.com/report/github.com/octohelm/jsonnetmod)

[JSONNET](https://jsonnet.org/) dependency management based on [go modules](https://golang.org/ref/mod)

## Requirements

* go 1.11+ (need `go mod download -json` to download & cache deps)

## Install

```shell
go get github.com/octohelm/jsonnetmod/cmd/jmod
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
    'github.com/rancher/local-path-provisioner': 'github.com/rancher/local-path-provisioner@v0.0.18',
    // local mod replace
    'github.com/x/a': '../a',
    // hack for k.libsonnet (force redirect to k/main.libsonnet)
    k: 'github.com/jsonnet-libs/k8s-alpha/1.19',
    // mod short alias
    'ksonnet-util': 'github.com/grafana/jsonnet-libs/ksonnet-util',
  },
  
  // automately resolve by the jsonnet code `import` or `importstr`
  // rules follow go modules
  // :: hidden fields means indirect require
  require: {
    'github.com/rancher/local-path-provisioner':: 'v0.0.19',
    'github.com/grafana/jsonnet-libs':: 'v0.0.0-20210209092858-49e80898b183',
    'github.com/x/a': 'latest',
  },
}
```

## Work with Tanka

```
jmod build -o ./environments/demo/main.jsonnet ./path/to/tanka-environment.jsonnet
tk show ./environments/demo
```

export not support pipe now, so need to create inline env object main.jsonnet first.


### Plugin kube

Like Tanka but without struct limit.
Only one requirement, make sure the file return [`tanka.dev/Environment` object](https://tanka.dev/inline-environments#converting-to-an-inline-environment)

```
cd ./examples
jmod k show ./clusters/demo/hello-world.jsonnet
jmod k apply ./clusters/demo/hello-world.jsonnet
jmod k delete ./clusters/demo/hello-world.jsonnet
jmod k prune ./clusters/demo/hello-world.jsonnet
```