package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/octohelm/tankax/pkg/tankax"
	"github.com/octohelm/tankax/pkg/terminal"
	"github.com/pkg/errors"
)

func init() {
	rootCmd.AddCommand(showCmd())
}

func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <path_to_jsonnet>",
		Short: "jsonnet as yaml",
	}

	wfFlows := workflowFlags(cmd.Flags())

	ouputDir := cmd.Flags().StringP("output-dir", "o", "", "write result to dir instead of stdout")

	showOrExport := func(jsonnetfile string, opts tankax.Opts) error {
		l, err := tankaxCtx.LoadClusterComponent(jsonnetfile, opts)
		if err != nil {
			return err
		}
		if *ouputDir != "" {
			for _, m := range l.Resources() {
				buf := bytes.Buffer{}

				if err := tmpl.Execute(&buf, m); err != nil {
					return errors.Wrapf(err, "executing name template")
				}

				name := strings.Replace(buf.String(), string(os.PathSeparator), "-", -1)

				env := l.Env()

				path := filepath.Join(*ouputDir, env.Metadata.Name, env.Spec.Namespace, name+".yaml")

				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					return fmt.Errorf("creating filepath '%s': %s", filepath.Dir(path), err)
				}

				if err := ioutil.WriteFile(path, []byte(m.String()), 0644); err != nil {
					return fmt.Errorf("writing manifest: %s", err)
				}

				fmt.Println("Written", path)
			}

			return nil
		}
		return terminal.Pageln(l.Resources().String())
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		for i := range args {
			opts := tankax.Opts{}
			opts.Filters = stringsToRegexps(wfFlows.targets)

			if err := showOrExport(args[i], opts); err != nil {
				return err
			}
		}

		return nil
	}

	return cmd
}

var tmpl, _ = template.New("").Parse("{{ .apiVersion }}.{{ .kind }}-{{ .metadata.name }}{{ if .metadata.namespace }}.{{ .metadata.namespace }}{{ end }}")
