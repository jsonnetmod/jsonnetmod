package vmod

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
)

func TestVendorMod(t *testing.T) {
	cwd, _ := os.Getwd()

	vm, err := VModFor(filepath.Join(cwd, "./testdata"))
	NewWithT(t).Expect(err).To(BeNil())

	zapLog, _ := zap.NewDevelopment()
	ctx := logr.NewContext(context.Background(), zapr.NewLogger(zapLog))

	t.Run("get", func(t *testing.T) {
		err := vm.Get(ctx, "github.com/rancher/local-path-provisioner/deploy/chart")
		NewWithT(t).Expect(err).To(BeNil())

		err = vm.Get(ctx, "github.com/rancher/local-path-provisioner/deploy/chart@v0.0.18")
		NewWithT(t).Expect(err).To(BeNil())

		err = vm.Get(ctx, "github.com/rancher/local-path-provisioner/deploy/chart")
		NewWithT(t).Expect(err).To(BeNil())

		err = vm.Get(ctx, "github.com/grafana/jsonnet-libs/ksonnet-util")
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("alias", func(t *testing.T) {
		err := vm.Alias(ctx, "local-path-provisioner", "github.com/rancher/local-path-provisioner/deploy/chart")
		NewWithT(t).Expect(err).To(BeNil())

		err = vm.Alias(ctx, "ksonnet-util", "github.com/grafana/jsonnet-libs/ksonnet-util")
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("download", func(t *testing.T) {
		err := vm.Download(ctx)
		NewWithT(t).Expect(err).To(BeNil())
	})
}
