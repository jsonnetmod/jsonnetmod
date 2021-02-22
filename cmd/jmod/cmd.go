package main

import (
	"github.com/go-logr/zapr"
	"github.com/octohelm/jsonnetmod/pkg/jsonnetmod"
	"github.com/octohelm/jsonnetmod/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	mod       *jsonnetmod.VMod
	zapLog, _ = zap.NewDevelopment(zap.WithCaller(false), zap.IncreaseLevel(zap.InfoLevel))
	log       = zapr.NewLogger(zapLog)
	rootCmd   = cmdRoot()
)

type ProjectOpts struct {
	Root    string `name:"project,p" usage:"project root dir"`
	Verbose bool   `name:"verbose,v" usage:"verbose"`
}

func cmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jmod",
		Short:   "jsonnet mod",
		Version: version.Version,
	}

	opts := ProjectOpts{Root: "."}

	return setupPersistentPreRun(cmd, &opts, func(cmd *cobra.Command, args []string) error {
		mod = jsonnetmod.VModFor(opts.Root)

		if opts.Verbose {
			zapLog, _ := zap.NewDevelopment()
			v := log.(zapr.Underlier).GetUnderlying()
			*v = *zapLog
		}

		return nil
	})
}
