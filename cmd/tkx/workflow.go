package main

import (
	"github.com/grafana/tanka/pkg/process"
	"github.com/spf13/pflag"
)

type workflowFlagVars struct {
	targets []string
}

func workflowFlags(fs *pflag.FlagSet) *workflowFlagVars {
	v := workflowFlagVars{}
	fs.StringSliceVarP(&v.targets, "target", "t", nil, "only use the specified objects (Format: <type>/<name>)")
	return &v
}

func stringsToRegexps(exps []string) process.Matchers {
	regexs, err := process.StrExps(exps...)
	if err != nil {
		log.Error(err, "")
	}
	return regexs
}
