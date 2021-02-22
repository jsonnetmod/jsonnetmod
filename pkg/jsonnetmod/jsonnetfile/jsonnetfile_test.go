package jsonnetfile

import (
	"testing"

	"github.com/octohelm/jsonnetmod/pkg/jsonnetmod/modfile"

	. "github.com/onsi/gomega"
)

func TestLoadModFile(t *testing.T) {
	mf := &modfile.ModFile{}

	_, err := LoadModFile("./testdata/kube-prometheus", mf)

	NewWithT(t).Expect(err).To(BeNil())
}
