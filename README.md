# Grafana Tanka Extra

A extra tool based on [Grafana Tanka](https://tanka.dev/)

## Code Structure

```
clusters/
    <cluster>/
        .cluster.json # config of spec.json
        <component>.jsonnet
```

## Commands

## Workflow

```bash 
tkx show ./clusters/demo/hello-world
tkx apply ./clusters/demo/hello-world 
tkx prune ./clusters/demo/hello-world 
tkx delete ./clusters/demo/hello-world
```

## PKG

## How to work

always run `main.jsonnet` on project root with code below:

```jsonnet
(import './clusters/<cluster>/<component>.jsonnet') + 
({
    _config+:: {
        name: 'xxx',
        namespace: '<component>'
    }    
})
```