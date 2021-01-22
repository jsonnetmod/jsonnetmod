package main

import (
	"context"

	"github.com/octohelm/tankax/pkg/jsonnetmod"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		getCmd(),
	)
}

func getCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "vmod",
	}

	o := jsonnetmod.VModOpts{}

	cmd.Flags().BoolVarP(&o.Upgrade, "upgrade", "u", false, "upgrade")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := logr.NewContext(context.Background(), log)

		importPath := "."
		if len(args) > 0 {
			importPath = args[0]
		}

		return vmod.Get(jsonnetmod.WithVModOpts(o)(ctx), importPath)
	}

	return cmd
}
