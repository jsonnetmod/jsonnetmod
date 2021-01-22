# Grafana Tanka Extra

[![GoDoc Widget](https://godoc.org/github.com/octohelm/tankax?status.svg)](https://godoc.org/github.com/octohelm/tankax)
[![codecov](https://codecov.io/gh/octohelm/tankax/branch/master/graph/badge.svg)](https://codecov.io/gh/octohelm/tankax)
[![Go Report Card](https://goreportcard.com/badge/github.com/octohelm/tankax)](https://goreportcard.com/report/github.com/octohelm/tankax)

An extra tool based on [Grafana Tanka](https://tanka.dev/)

## Requirements

* go 1.15+ (need `go mod download -json` to download & cache deps)

## Install

```
go get github.com/octohelm/tankax/cmd/tkx
```

## Code Structure

```
clusters/
    <cluster>/
        .cluster.json # config of spec.json
        <namespace>[.<component>].jsonnet
mod.jsonnet
```

## Features

### Jsonnet Mod

No vendor, same as go mod, all deps will download as go mod cache.

```shell
# auto dep
tkx get ./...

# upgrade dep
tkx get -u ./...

# install dep with special version
tkx get github.com/grafana/jsonnet-libs@latest
```

### Workflow

```shell 
tkx show ./clusters/demo/hello-world.jsonnet
tkx apply ./clusters/demo/hello-world.jsonnet 
tkx prune ./clusters/demo/hello-world.jsonnet 
tkx delete ./clusters/demo/hello-world.jsonnet
```

## How to work

always run a fake `main.jsonnet` on project root with code below:

```jsonnet
(import './clusters/<cluster>/<namespace>.<component>.jsonnet') + 
({
    _config+:: {
        namespace: '<namespace>',
    }    
})
```

### Helm

Tankax support import yaml, so helm could be use by

```jsonnet
local c = import 'github.com/rancher/local-path-provisioner/deploy/chart/Chart.yaml';

{
  _chart:: c,
  _values+:: {},
}
```