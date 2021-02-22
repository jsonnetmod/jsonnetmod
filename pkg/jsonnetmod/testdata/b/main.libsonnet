local k = import 'ksonnet-util/kausal.libsonnet';
local grafana = import 'github.com/grafana/jsonnet-libs/grafana/grafana.libsonnet';

(import 'github.com/x/a/main.libsonnet') +  grafana  + {
    namespace:
       k.core.v1.namespace.new("test"),
}