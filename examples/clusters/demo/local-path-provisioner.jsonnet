local demo = import 'examples/clusters/demo.jsonnet';
local data = (import 'examples/components/local-path-provisioner/main.jsonnet').helm();

demo(data, 'local-path-provisioner')
