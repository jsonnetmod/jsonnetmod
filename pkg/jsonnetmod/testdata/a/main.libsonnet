local c = importstr 'github.com/rancher/local-path-provisioner/deploy/chart/Chart.yaml';

{
  _config+:: {
    name: c.name,
  },

  _chart:: c,
  _values+:: {}
}
