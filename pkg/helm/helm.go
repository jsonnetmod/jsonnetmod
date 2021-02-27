package helm

import (
	"io"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chart"

	"github.com/grafana/tanka/pkg/helm"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type TemplateOnlyHelm struct {
}

func (TemplateOnlyHelm) Pull(chart, version string, opts helm.PullOpts) error {
	panic("template only")
}

func (TemplateOnlyHelm) RepoUpdate(opts helm.Opts) error {
	panic("template only")
}

func (TemplateOnlyHelm) Template(name, chart string, opts helm.TemplateOpts) (manifest.List, error) {
	c, err := loader.LoadDir(chart)
	if err != nil {
		return nil, errors.Wrapf(err, "load helm chart failed from %s", chart)
	}

	renderedContentMap, err := HelmTemplate(c, opts.Values, chartutil.ReleaseOptions{
		Name:      name,
		Namespace: opts.Namespace,
	})
	if err != nil {
		return nil, errors.Wrap(err, "render helm failed")
	}

	var list manifest.List

	if opts.IncludeCRDs {
		for _, c := range c.CRDObjects() {
			renderedContentMap[c.Filename] = string(c.File.Data)
		}
	}

	for fileName, renderedContent := range renderedContentMap {
		if filepath.Ext(fileName) != ".yaml" || filepath.Ext(fileName) == ".yml" {
			continue
		}

		if strings.TrimSpace(renderedContent) != "" {
			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(renderedContent), 4096)

			var m manifest.Manifest

			if err := decoder.Decode(&m); err != nil {
				if err == io.EOF {
					break
				}
				return nil, errors.Wrap(err, "helm template failed")
			}

			if len(m) == 0 {
				continue
			}

			list = append(list, m)
		}

	}

	return list, nil
}

func HelmTemplate(c *chart.Chart, values map[string]interface{}, opts chartutil.ReleaseOptions) (map[string]string, error) {
	values, err := chartutil.CoalesceValues(c, values)
	if err != nil {
		return nil, err
	}

	valuesToRender, err := chartutil.ToRenderValues(c, values, opts, nil)
	if err != nil {
		return nil, err
	}

	e := &engine.Engine{}
	return e.Render(c, valuesToRender)
}
