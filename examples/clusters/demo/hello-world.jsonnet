local demo = import '../demo.jsonnet';
local data = (import 'examples/components/hello-world/main.jsonnet') +
             {
               _config+:: {
                 name: 'hello-world-test',
                 namespace: 'hello-world',
               },
             };

demo(data, data._config.namespace)
