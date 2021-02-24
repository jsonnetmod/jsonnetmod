package forkinternal

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestForkInternals(t *testing.T) {
	err := ForkInternals("../../forked", "cmd/go/internal/modload", "cmd/go/internal/modfetch")
	NewWithT(t).Expect(err).To(BeNil())
}
