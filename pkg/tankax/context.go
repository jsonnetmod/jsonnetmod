package tankax

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/octohelm/tankax/pkg/jsonnetmod"

	"github.com/grafana/tanka/pkg/jsonnet/jpath"
	"github.com/grafana/tanka/pkg/process"
	"github.com/grafana/tanka/pkg/tanka"
	"github.com/pkg/errors"
)

func NewContext(ctx context.Context, vmod *jsonnetmod.VMod) *Context {
	c := &Context{vmod: vmod, ctx: ctx}

	c.CWD, _ = os.Getwd()

	return c
}

type Context struct {
	ctx      context.Context
	CWD      string
	vmod     *jsonnetmod.VMod
	Clusters map[string]Cluster
}

type Cluster struct {
	Server string
}

func (c *Context) resolvePath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(c.CWD, filename)
}

func (c *Context) LoadClusterComponent(filename string, opts tanka.Opts) (*tanka.LoadResult, error) {
	env, err := c.parseEnv(c.resolvePath(filename))
	if err != nil {
		return nil, err
	}

	rec, err := process.Process(*env, opts.Filters)
	if err != nil {
		return nil, err
	}

	return &tanka.LoadResult{
		Env:       env,
		Resources: rec,
	}, nil
}

func (c *Context) parseEnv(jsonnetAbsFile string) (*Environment, error) {
	dir := filepath.Dir(jsonnetAbsFile)

	jsonnetRoot := filepath.Join(c.vmod.Dir, jpath.DEFAULT_ENTRYPOINT)

	clusterJSON := filepath.Join(dir, ".cluster.json")

	data, err := ioutil.ReadFile(clusterJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "missing %s", clusterJSON)
	}
	spec := Spec{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling data")
	}

	// ./<cluster_name>/<namespace>.xxx.jsonnet
	clusterName := lastDirname(jsonnetAbsFile)
	namespace := strings.TrimSuffix(filepath.Base(jsonnetAbsFile), filepath.Ext(jsonnetAbsFile))

	env := &Environment{}

	env.Metadata.Name = clusterName

	env.Spec = spec
	env.Spec.InjectLabels = true
	env.Spec.Namespace = namespace

	evalScript := fmt.Sprintf(`
(import '%s') + 
{
  _config+:: {
    namespace: '%s',
  },

  _chart+:: {},	
  _values+:: {},	
  
  helm: (
	if (
		std.objectHasAll($._chart, "__dirname") && std.isString($._chart.__dirname)
		&& 
		std.objectHasAll($._chart, "name") && std.isString($._chart.name)
	)
	then std.native('helmTemplate')(
      if (std.objectHasAll($._config, "name") && std.isString($._config.name)) then $._config.name else $._chart.name, 
      $._chart.__dirname, 
      $._values + { calledFrom: '/' },
    )
	else {} 
  ),	
}
`, jsonnetAbsFile, env.Spec.Namespace)

	if err := c.evaluate(&env.Data, jsonnetRoot, evalScript); err != nil {
		return nil, err
	}

	return env, nil
}

func (c *Context) evaluate(v interface{}, file string, evalScript string) error {
	vm := c.vmod.MakeVM(c.ctx)
	for _, nf := range native.Funcs() {
		vm.NativeFunction(nf)
	}

	raw, err := vm.EvaluateAnonymousSnippet(file, evalScript)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(raw), v)
}
