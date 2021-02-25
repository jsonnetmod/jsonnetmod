package main

import (
	"context"

	"github.com/jsonnetmod/jsonnetmod/pkg/jsonnetmod"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		cmdGet(),
	)
}

func cmdGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "download dependencies",
	}

	o := jsonnetmod.Opts{}

	return setupRun(cmd, &o, func(ctx context.Context, args []string) error {
		importPath := "."
		if len(args) > 0 {
			importPath = args[0]
		}
		return mod.Get(jsonnetmod.WithOpts(ctx, jsonnetmod.OptUpgrade(o.Upgrade)), importPath)
	})
}
