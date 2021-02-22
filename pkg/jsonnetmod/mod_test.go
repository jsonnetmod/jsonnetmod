package jsonnetmod

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMod(t *testing.T) {

	t.Run("load", func(t *testing.T) {
		mod, err := ModFromDir("./testdata/b")

		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(mod.Module).To(Equal("github.com/x/b"))

		NewWithT(t).Expect(mod.Replace[PathIdentity{
			Path: "github.com/rancher/local-path-provisioner",
		}]).To(Equal(PathIdentity{
			Path:    "github.com/rancher/local-path-provisioner",
			Version: "v0.0.18",
		}))
	})
}
