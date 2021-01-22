local k = import 'ksonnet-util/kausal.libsonnet';
local c = import 'github.com/rancher/local-path-provisioner/deploy/chart/Chart.yaml';

{
  _config+:: {
    name: c.name,
  },

  _chart:: c,
  _values+:: {
    nodeSelector: {
      'node-role.kubernetes.io/master': 'true',
    },
    tolerations: [
      {
        key: 'node-role.kubernetes.io/master',
        operator: 'Exists',
        effect: 'NoSchedule',
      },
    ],
    nodePathMap: [
      {
        node: 'DEFAULT_PATH_FOR_NON_LISTED_NODES',
        paths: [
          '/data/local-path-provisioner',
        ],
      },
    ],
    configmap: {
      helperPod: std.manifestYamlDoc({
        apiVersion: 'v1',
        kind: 'Pod',
        metadata: {
          name: 'helper-pod',
        },
        spec: {
          containers: [
            {
              name: 'helper-pod',
              image: 'docker.io/library/busybox',
            },
          ],
        },
      }),
    },
  },
}
