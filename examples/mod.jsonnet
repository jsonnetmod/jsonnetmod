{
  module: 'examples',
  jpath: './vendor',
  replace: {
    k: 'github.com/jsonnet-libs/k8s-alpha/1.19',
    'ksonnet-util': 'github.com/grafana/jsonnet-libs/ksonnet-util',
  },
  require: {
    'github.com/grafana/jsonnet-libs': 'v0.0.0-20210209092858-49e80898b183',
    'github.com/jsonnet-libs/k8s-alpha':: 'v0.0.0-20210118111845-5e0d0738721f',
    'github.com/rancher/local-path-provisioner': 'v0.0.19',
  },
}
