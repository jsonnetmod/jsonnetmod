package tankax

import (
	"github.com/grafana/tanka/pkg/kubernetes"
	"github.com/grafana/tanka/pkg/spec/v1alpha1"
	"github.com/grafana/tanka/pkg/tanka"
)

type Loaded interface {
	Connect() (*kubernetes.Kubernetes, error)
	Result() *tanka.LoadResult
}

type Spec = v1alpha1.Spec
type Opts = tanka.Opts
type Environment = v1alpha1.Environment
