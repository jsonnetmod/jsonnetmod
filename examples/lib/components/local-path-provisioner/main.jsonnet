local k = import 'ksonnet-util/kausal.libsonnet';

{
  _config+:: {
    name: 'local-path-provisioner',
    chart: './vendor/local-path-provisioner',
  },

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
      helperPod: k.util.manifestYaml({
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
