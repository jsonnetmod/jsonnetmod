package jsonnetmod

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func TestModCache(t *testing.T) {

	zapLog, _ := zap.NewDevelopment()
	ctx := logr.NewContext(context.Background(), zapr.NewLogger(zapLog))

	m := NewModCache()

	t.Run("should resolve repo root", func(t *testing.T) {
		repo := "github.com/kubernetes-monitoring/kubernetes-mixin/lib/promgrafonnet"

		resolveRepo, err := m.repoRoot(ctx, repo)

		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(resolveRepo).To(Equal("github.com/kubernetes-monitoring/kubernetes-mixin"))
	})

	t.Run("should resolve sub dir with go.mod", func(t *testing.T) {
		repo := "github.com/prometheus/node_exporter/docs/node-mixin"

		resolveRepo, err := m.repoRoot(ctx, repo)

		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(resolveRepo).To(Equal(repo))
	})
}
