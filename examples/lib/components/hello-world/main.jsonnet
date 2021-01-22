local k = import 'github.com/grafana/jsonnet-libs/ksonnet-util/kausal.libsonnet';

{
  _config+:: {
    name: 'nginx',
    namespace: 'default',
  },

  _images+:: {
    nginx: 'nginx:latest',
  },

  namespace:
    k.core.v1.namespace.new($._config.namespace),

  deployment:
    k.apps.v1.deployment.new(
      name=$._config.name,
      replicas=1,
      containers=[
        k.core.v1.container.new(
          name='nginx',
          image=$._images.nginx,
        )
        + k.core.v1.container.withPorts([
          k.core.v1.containerPort.new(
            name='http',
            port=80,
          ),
        ]),
      ]
    ),

  svc:
    k.util.serviceFor($.deployment),
}
