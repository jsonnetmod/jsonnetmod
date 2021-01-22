package main

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/grafana/tanka/pkg/tanka"
	"github.com/octohelm/tankax/pkg/tankax"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pruneCmd())
}

func pruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune <path_to_jsonnet>",
		Short: "prune the environment from cluster",
	}

	var opts tanka.PruneOpts

	cmd.Flags().BoolVar(&opts.Force, "force", false, "force deleting (kubectl prune --force)")
	cmd.Flags().BoolVar(&opts.AutoApprove, "dangerous-auto-approve", false, "skip interactive approval. Only for automation!")

	vars := workflowFlags(cmd.Flags())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := logr.NewContext(context.Background(), log)

		opts.Filters = stringsToRegexps(vars.targets)

		loaded, err := tankax.NewContext(ctx, vmod).LoadClusterComponent(args[0], opts.Opts)
		if err != nil {
			return err
		}

		return tankax.Prune(loaded, opts)
	}
	return cmd
}
