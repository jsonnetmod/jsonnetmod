package main

import (
	"context"

	"github.com/go-logr/zapr"
	"github.com/jsonnetmod/jsonnetmod/pkg/jsonnetmod"
	"github.com/jsonnetmod/jsonnetmod/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	mod            *jsonnetmod.VMod
	zapLog, _      = zap.NewDevelopment(zap.WithCaller(false), zap.IncreaseLevel(zap.InfoLevel))
	log            = zapr.NewLogger(zapLog)
	projectOptions = ProjectOpts{Root: "."}
	rootCmd        = cmdRoot()
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

	return setupPersistentPreRun(cmd, &projectOptions, func(ctx context.Context, args []string) error {
		mod = jsonnetmod.VModFor(projectOptions.Root)

		if projectOptions.Verbose {
			zapLog, _ := zap.NewDevelopment()
			v := log.(zapr.Underlier).GetUnderlying()
			*v = *zapLog
		}

		return nil
	})
}
