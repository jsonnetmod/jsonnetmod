function(data={}, namespace='default') {
  apiVersion: 'tanka.dev/v1alpha1',
  kind: 'Environment',
  metadata: {
    name: 'demo',
  },
  spec: {
    namespace: namespace,
    apiServer: 'https://172.16.0.7:8443',
    injectLabels: true,
  },
  data: data,
}
