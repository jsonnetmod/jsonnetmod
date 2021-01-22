package main

import (
	"fmt"
	"io"
	"os"

	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/octohelm/jsonnetmod/pkg/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		cmdBuild(),
	)
}

type BuildOpts struct {
	Output string `name:"output,o" usage:"output filename"`
}

func cmdBuild() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build <input>",
		Short: "build to json",
	}

	opts := BuildOpts{}

	return setupRun(cmd, &opts, func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing input")
		}

		vm := mod.MakeVM(cmd.Context())

		for _, nf := range native.Funcs() {
			vm.NativeFunction(nf)
		}

		jsonData, err := vm.EvaluateFile(args[0])
		if err != nil {
			return err
		}

		if o := opts.Output; o != "" {
			if err := util.WriteFile(o, []byte(jsonData)); err != nil {
				return err
			}
		} else {
			_, _ = io.WriteString(os.Stdout, jsonData)
		}

		return nil
	})
}
