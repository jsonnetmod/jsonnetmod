package main

import (
	"github.com/octohelm/jsonnetmod/pkg/jsonnet"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		cmdFmt(),
	)
}

func cmdFmt() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "format jsonnet codes",
	}

	formatOpts := jsonnet.FormatOpts{}

	return setupRun(cmd, &formatOpts, func(cmd *cobra.Command, args []string) error {
		baseDir := "./"
		if len(args) > 0 {
			baseDir = args[0]
		}

		files, err := mod.ListJsonnet(baseDir)
		if err != nil {
			return err
		}

		return jsonnet.FormatFiles(cmd.Context(), files, formatOpts)
	})
}
