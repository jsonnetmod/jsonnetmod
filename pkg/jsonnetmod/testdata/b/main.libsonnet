local k = import 'ksonnet-util/kausal.libsonnet';

(import 'github.com/x/a/main.libsonnet') + {
    namespace:
       k.core.v1.namespace.new("test"),
}