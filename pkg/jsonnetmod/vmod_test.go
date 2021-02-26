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

	ctx = WithOpts(ctx, OptVerbose(true))

	t.Run("mod a", func(t *testing.T) {
		a := VModFor(filepath.Join(cwd, "./testdata/a"))

		_ = os.RemoveAll(filepath.Join(a.Dir, "vendor"))

		t.Run("get", func(t *testing.T) {
			err := a.Get(ctx, ".")
			NewWithT(t).Expect(err).To(BeNil())

			NewWithT(t).Expect(a.Mod.Require["github.com/rancher/local-path-provisioner"].Version).To(Equal("v0.0.19"))
			NewWithT(t).Expect(VendorSymlink(a.Mod, "github.com/rancher/local-path-provisioner")).To(HaveSuffix("github.com/rancher/local-path-provisioner@v0.0.19"))

		})
	})

	t.Run("mod b", func(t *testing.T) {
		b := VModFor(filepath.Join(cwd, "./testdata/b"))
		_ = os.RemoveAll(filepath.Join(b.Dir, "vendor"))

		t.Run("get with auto imports", func(t *testing.T) {
			err := b.Get(ctx, ".")
			NewWithT(t).Expect(err).To(BeNil())

			NewWithT(t).Expect(b.Mod.Require["github.com/rancher/local-path-provisioner"].Version).To(Equal("v0.0.19"))
			NewWithT(t).Expect(b.Mod.Require["github.com/rancher/local-path-provisioner"].Indirect).To(Equal(true))
			NewWithT(t).Expect(VendorSymlink(b.Mod, "github.com/rancher/local-path-provisioner")).To(HaveSuffix("github.com/rancher/local-path-provisioner@v0.0.18"))
		})

		t.Run("get with upgrade", func(t *testing.T) {
			err := b.Get(WithOpts(ctx, OptUpgrade(true)), ".")
			NewWithT(t).Expect(err).To(BeNil())

			NewWithT(t).Expect(b.Mod.Require["github.com/rancher/local-path-provisioner"].Version).To(Equal("v0.0.19"))
			NewWithT(t).Expect(b.Mod.Require["github.com/rancher/local-path-provisioner"].Indirect).To(Equal(true))
			NewWithT(t).Expect(VendorSymlink(b.Mod, "github.com/rancher/local-path-provisioner")).To(HaveSuffix("github.com/rancher/local-path-provisioner@v0.0.18"))
		})
	})

	t.Run("mod c", func(t *testing.T) {
		c := VModFor(filepath.Join(cwd, "./testdata/c"))
		_ = os.RemoveAll(filepath.Join(c.Dir, "vendor"))

		t.Run("get with auto imports", func(t *testing.T) {
			err := c.Get(ctx, ".")
			NewWithT(t).Expect(err).To(BeNil())

			NewWithT(t).Expect(c.Mod.Require["github.com/rancher/local-path-provisioner"].Version).To(Equal("v0.0.19"))
			NewWithT(t).Expect(c.Mod.Require["github.com/rancher/local-path-provisioner"].Indirect).To(Equal(true))
			NewWithT(t).Expect(VendorSymlink(c.Mod, "github.com/rancher/local-path-provisioner")).To(HaveSuffix("github.com/rancher/local-path-provisioner@v0.0.18"))
		})
	})
}

func VendorSymlink(m *Mod, module string) string {
	p, _ := os.Readlink(filepath.Join(m.Dir, "./vendor/", module))
	return p
}
