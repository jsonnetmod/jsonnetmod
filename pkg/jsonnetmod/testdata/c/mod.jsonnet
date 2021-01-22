{
  module: 'github.com/x/c',
  jpath: './vendor',
  replace: {
    'github.com/x/b': '../b',
  },
  require: {
    'github.com/grafana/jsonnet-libs':: 'v0.0.0-20210219085724-4db296c6a39e',
    'github.com/jsonnet-libs/docsonnet':: 'v0.0.0-20200817072722-3e1757637edf',
    'github.com/jsonnet-libs/k8s-alpha':: 'v0.0.0-20210118111845-5e0d0738721f',
    'github.com/rancher/local-path-provisioner':: 'v0.0.19',
    'github.com/x/a':: 'latest',
    'github.com/x/b': 'latest',
  },
}
