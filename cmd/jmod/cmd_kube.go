package main

import (
	"context"
	"fmt"

	"github.com/grafana/tanka/pkg/jsonnet/native"
	"github.com/octohelm/jsonnetmod/pkg/tanka"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		cmdKube(),
	)
}

func cmdKube() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kube",
		Aliases: []string{"k"},
	}

	cmd.AddCommand(
		cmdKubeApply(),
		cmdKubeShow(),
		cmdKubeDelete(),
		cmdKubePrune(),
	)

	return cmd
}

func cmdKubeShow() *cobra.Command {
	cmd := &cobra.Command{
		Use: "show <input>",
	}

	opts := tanka.ShowOpts{}

	return setupRun(cmd, &opts, func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing input")
		}

		lr, err := load(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		return lr.Show(opts)
	})
}

func cmdKubeApply() *cobra.Command {
	cmd := &cobra.Command{
		Use: "apply <input>",
	}

	opts := tanka.ApplyOpts{
		Validate: true,
	}

	return setupRun(cmd, &opts, func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing input")
		}

		lr, err := load(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		return lr.Apply(opts)
	})
}

func cmdKubeDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete <input>",
	}

	opts := tanka.DeleteOpts{
		Validate: true,
	}

	return setupRun(cmd, &opts, func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing input")
		}

		lr, err := load(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		return lr.Delete(opts)
	})
}

func cmdKubePrune() *cobra.Command {
	cmd := &cobra.Command{
		Use: "prune <input>",
	}

	opts := tanka.PruneOpts{}

	return setupRun(cmd, &opts, func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing input")
		}

		lr, err := load(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		return lr.Prune(opts)
	})
}

func load(ctx context.Context, filename string) (*tanka.LoadResult, error) {
	vm := mod.MakeVM(ctx)

	for _, nf := range native.Funcs() {
		vm.NativeFunction(nf)
	}

	jsonData, err := vm.EvaluateFile(filename)
	if err != nil {
		return nil, err
	}

	return tanka.Process([]byte(jsonData), nil)
}
