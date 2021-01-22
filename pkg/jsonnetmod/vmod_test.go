package jsonnetmod

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

	zapLog, _ := zap.NewDevelopment()
	ctx := logr.NewContext(context.Background(), zapr.NewLogger(zapLog))

	t.Run("mod a", func(t *testing.T) {
		a := VModFor(filepath.Join(cwd, "./testdata/a"))

		_ = os.RemoveAll(filepath.Join(a.Dir, "vendor"))

		t.Run("get", func(t *testing.T) {
			err := a.Get(ctx, ".")
			NewWithT(t).Expect(err).To(BeNil())

			p, err := os.Readlink(filepath.Join(a.Dir, "./vendor/github.com/rancher/local-path-provisioner"))
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(p).To(HaveSuffix("github.com/rancher/local-path-provisioner@v0.0.19"))
		})
	})

	t.Run("mod b", func(t *testing.T) {
		b := VModFor(filepath.Join(cwd, "./testdata/b"))
		_ = os.RemoveAll(filepath.Join(b.Dir, "vendor"))

		t.Run("get with auto imports", func(t *testing.T) {
			err := b.Get(ctx, ".")
			NewWithT(t).Expect(err).To(BeNil())

			p, err := os.Readlink(filepath.Join(b.Dir, "./vendor/github.com/rancher/local-path-provisioner"))
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(p).To(HaveSuffix("github.com/rancher/local-path-provisioner@v0.0.18"))
		})

		t.Run("get with upgrade", func(t *testing.T) {
			err := b.Get(WithVModOpts(VModOpts{Upgrade: true})(ctx), ".")
			NewWithT(t).Expect(err).To(BeNil())

			p, err := os.Readlink(filepath.Join(b.Dir, "./vendor/github.com/rancher/local-path-provisioner"))
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(p).To(HaveSuffix("github.com/rancher/local-path-provisioner@v0.0.18"))
		})
	})
}
