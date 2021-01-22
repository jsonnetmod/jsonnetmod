package main

import (
	"github.com/grafana/tanka/pkg/tanka"
	"github.com/octohelm/tankax/pkg/tankax"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(applyCmd())
}

func applyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply <path_to_jsonnet>",
		Short: "apply the configuration to the cluster",
	}

	var opts tanka.ApplyOpts

	cmd.Flags().BoolVar(&opts.Force, "force", false, "force applying (kubectl apply --force)")
	cmd.Flags().BoolVar(&opts.Validate, "validate", true, "validation of resources (kubectl --validate=false)")
	cmd.Flags().BoolVar(&opts.AutoApprove, "dangerous-auto-approve", false, "skip interactive approval. Only for automation!")

	vars := workflowFlags(cmd.Flags())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts.Filters = stringsToRegexps(vars.targets)

		loaded, err := tankaxCtx.LoadClusterComponent(args[0], opts.Opts)
		if err != nil {
			return err
		}

		return tankax.Apply(loaded, opts)
	}

	return cmd
}
