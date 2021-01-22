package main

import (
	"github.com/grafana/tanka/pkg/tanka"
	"github.com/octohelm/tankax/pkg/tankax"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd())
}

func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <path_to_jsonnet>",
		Short: "delete the environment from cluster",
	}

	var opts tanka.DeleteOpts

	cmd.Flags().BoolVar(&opts.Force, "force", false, "force deleting (kubectl delete --force)")
	cmd.Flags().BoolVar(&opts.Validate, "validate", true, "validation of resources (kubectl --validate=false)")
	cmd.Flags().BoolVar(&opts.AutoApprove, "dangerous-auto-approve", false, "skip interactive approval. Only for automation!")

	vars := workflowFlags(cmd.Flags())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts.Filters = stringsToRegexps(vars.targets)

		loaded, err := tankaxCtx.LoadClusterComponent(args[0], opts.Opts)
		if err != nil {
			return err
		}

		return tankax.Delete(loaded, opts)
	}
	return cmd
}
