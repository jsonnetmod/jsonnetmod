package tankax

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/tanka/pkg/jsonnet"
	"github.com/grafana/tanka/pkg/jsonnet/jpath"
	"github.com/grafana/tanka/pkg/kubernetes"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/grafana/tanka/pkg/process"
	"github.com/grafana/tanka/pkg/tanka"
	"github.com/pkg/errors"
)

func NewContext(opt ContextOpts) *Context {
	c := &Context{ContextOpts: opt}

	c.CWD, _ = os.Getwd()

	return c
}

type ContextOpts struct {
	ProjectRoot string
	JsonnetHome []string
}

type Context struct {
	ContextOpts
	Clusters map[string]Cluster
	CWD      string
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

func (c *Context) LoadClusterComponent(filename string, opts tanka.Opts) (Loaded, error) {
	env, err := c.parseEnv(c.resolvePath(filename), opts.JsonnetOpts)
	if err != nil {
		return nil, err
	}

	rec, err := process.Process(*env, opts.Filters)
	if err != nil {
		return nil, err
	}

	return &loaded{
		resources: rec,
		env:       env,
	}, nil
}

func (c *Context) parseEnv(jsonnetAbsFile string, jsonnetOpts tanka.JsonnetOpts) (*Environment, error) {
	dir := filepath.Dir(jsonnetAbsFile)

	jsonnetRoot := filepath.Join(c.ProjectRoot, jpath.DEFAULT_ENTRYPOINT)

	jsonnetOpts.ImportPaths = make([]string, len(c.JsonnetHome))
	for i := range jsonnetOpts.ImportPaths {
		jsonnetOpts.ImportPaths[i] = filepath.Join(c.ProjectRoot, c.JsonnetHome[i])
	}

	clusterJSON := filepath.Join(dir, ".cluster.json")

	data, err := ioutil.ReadFile(clusterJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "missing %s", clusterJSON)
	}
	spec := Spec{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling data")
	}

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

  _values+:: {},	
  
  _helm: (
	if (
		std.objectHas($._config, "name") && std.isString($._config.name)
		&&
		std.objectHas($._config, "chart") && std.isString($._config.chart)
	)
	then std.native('helmTemplate')($._config.name, $._config.chart, $._values + { calledFrom: '%s' })
	else {} 
  ),	
}
`, jsonnetAbsFile, env.Spec.Namespace, jsonnetRoot)

	if err := evaluate(&env.Data, jsonnetRoot, evalScript, jsonnetOpts); err != nil {
		return nil, err
	}

	return env, nil
}

func evaluate(v interface{}, file string, evalScript string, opts tanka.JsonnetOpts) error {
	vm := jsonnet.MakeVM(opts)
	raw, err := vm.EvaluateAnonymousSnippet(file, evalScript)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(raw), v)
}

type loaded struct {
	env       *Environment
	resources manifest.List
}

func (p *loaded) Resources() manifest.List {
	return p.resources
}

func (p *loaded) Env() *Environment {
	return p.env
}

// connect opens a connection to the backing Kubernetes cluster.
func (p *loaded) Connect() (*kubernetes.Kubernetes, error) {
	env := *p.env
	// connect client
	kube, err := kubernetes.New(env)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to Kubernetes")
	}

	return kube, nil
}
