package main

import (
	"context"

	"github.com/octohelm/tankax/pkg/vmod"

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

	upgrade := cmd.Flags().BoolP("upgrade", "u", false, "upgrade")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		vm, err := vmod.VModFor(tankaxCtx.ProjectRoot)
		if err != nil {
			return err
		}

		ctx := logr.NewContext(context.Background(), log)

		importPath := "."
		if len(args) > 0 {
			importPath = args[0]
		}

		if importPath[0] == '.' {
			// todo auto import
			if *upgrade {
				return vm.Upgrade(ctx)
			}
			return vm.Download(ctx)
		}
		return vm.Get(ctx, importPath)
	}

	return cmd
}
