{
  module: 'github.com/x/b',
  jpath: './vendor',
  replace: {
    'doc-util': 'github.com/jsonnet-libs/docsonnet/doc-util',
    'github.com/rancher/local-path-provisioner': 'github.com/rancher/local-path-provisioner@v0.0.18',
    'github.com/x/a': '../a',
    k: 'github.com/jsonnet-libs/k8s-alpha/1.18',
    'ksonnet-util': 'github.com/grafana/jsonnet-libs/ksonnet-util',
  },
  require: {
    'github.com/grafana/jsonnet-libs': 'v0.0.0-20210219085724-4db296c6a39e',
    'github.com/jsonnet-libs/docsonnet':: 'v0.0.0-20200817072722-3e1757637edf',
    'github.com/jsonnet-libs/k8s-alpha':: 'v0.0.0-20210118111845-5e0d0738721f',
    'github.com/rancher/local-path-provisioner':: 'v0.0.19',
    'github.com/x/a': 'latest',
  },
}
