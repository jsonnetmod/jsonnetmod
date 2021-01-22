{
  module: 'github.com/x/b',
  jpath: './vendor',
  replace: {
    'github.com/rancher/local-path-provisioner': 'github.com/rancher/local-path-provisioner@v0.0.18',
    'github.com/x/a': '../a',
    k: 'github.com/jsonnet-libs/k8s-alpha/1.19',
  },
  require: {
    'github.com/rancher/local-path-provisioner':: 'v0.0.19',
    'github.com/x/a': 'v0.0.0',
  },
}
