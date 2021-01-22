package main

import (
	"fmt"

	"github.com/octohelm/tankax/pkg/glob"
	"github.com/spf13/cobra"

	"github.com/grafana/tanka/pkg/tanka"
)

func init() {
	rootCmd.AddCommand(
		fmtCmd(),
	)
}

func fmtCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "format Jsonnet code",
	}

	writeToFile := cmd.Flags().BoolP("write", "w", false, "write result to (source) file instead of stdout")
	listFiles := cmd.Flags().BoolP("list", "l", false, "list files whose formatting differs from")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		baseDir := "./"
		if len(args) > 0 {
			baseDir = args[0]
		}

		files, err := glob.Src(
			[]string{
				"*.*sonnet",
				"**/*.*sonnet",
				"!vendor/**",
				"!**/vendor/**",
			},
			glob.Dir(baseDir),
		)
		if err != nil {
			return err
		}

		changed, err := tanka.FormatFiles(files, &tanka.FormatOpts{
			PrintNames: *listFiles,
			OutFn:      toOutFn(*writeToFile),
		})
		if err != nil {
			return err
		}

		switch {
		case len(changed) > 0:
			log.Info("The following files are not properly formatted:")
			for _, s := range changed {
				log.Info(s)
			}
		case len(changed) == 0:
			log.Info("All discovered files are already formatted. No changes were made")
		case len(changed) > 0:
			log.Info(fmt.Sprintf("Formatted %v files", len(changed)))
		}

		return nil
	}

	return cmd
}

func toOutFn(writeToFile bool) func(name, content string) error {
	if writeToFile {
		return nil
	}

	return func(name, content string) error {
		fmt.Printf("// %s\n%s\n", name, content)
		return nil
	}
}
