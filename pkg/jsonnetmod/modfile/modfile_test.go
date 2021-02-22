package modfile

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestLoadModFile(t *testing.T) {
	mod := ModFile{}

	_, err := LoadModFile("../testdata/b", &mod)

	NewWithT(t).Expect(err).To(BeNil())
	NewWithT(t).Expect(mod.Module).To(Equal("github.com/x/b"))

	NewWithT(t).Expect(mod.Replace[PathIdentity{
		Path: "github.com/rancher/local-path-provisioner",
	}]).To(Equal(PathIdentity{
		Path:    "github.com/rancher/local-path-provisioner",
		Version: "v0.0.18",
	}))
}
