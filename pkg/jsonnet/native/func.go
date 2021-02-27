package native

import (
	"github.com/google/go-jsonnet"
	tankahelm "github.com/grafana/tanka/pkg/helm"
	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/jsonnetmod/jsonnetmod/pkg/helm"
)

func Funcs() []*jsonnet.NativeFunction {
	funcs := native.Funcs()

	for i, f := range funcs {
		if f.Name == "helmTemplate" {
			funcs[i] = tankahelm.NativeFunc(&helm.TemplateOnlyHelm{})
		}
	}

	return funcs
}
