package main

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/grafana/tanka/pkg/tanka"
	"github.com/octohelm/tankax/pkg/jsonnetmod"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	vmod *jsonnetmod.VMod

	log logr.Logger

	rootCmd = cmdRoot()
)

func cmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tkx",
		Short:   "tanka extra",
		Version: tanka.CURRENT_VERSION,
	}

	projectRoot := cmd.PersistentFlags().StringP("project", "p", ".", "project root")
	verbose := cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose")

	zapLog, _ := zap.NewDevelopment(zap.WithCaller(false), zap.IncreaseLevel(zap.InfoLevel))
	log = zapr.NewLogger(zapLog)

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		vmod = jsonnetmod.VModFor(*projectRoot)

		if *verbose {
			zapLog, _ := zap.NewDevelopment()
			log = zapr.NewLogger(zapLog)
		}
	}

	return cmd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err, "execute failed")
	}
}
