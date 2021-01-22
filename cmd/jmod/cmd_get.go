package main

import (
	"github.com/octohelm/jsonnetmod/pkg/jsonnetmod"

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

	return setupRun(cmd, &o, func(cmd *cobra.Command, args []string) error {
		importPath := "."
		if len(args) > 0 {
			importPath = args[0]
		}
		return mod.Get(jsonnetmod.WithOpts(cmd.Context(), jsonnetmod.OptUpgrade(o.Upgrade)), importPath)
	})
}
