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

	t.Run("should resolve repo repoRoot", func(t *testing.T) {
		repo := "github.com/kubernetes-monitoring/kubernetes-mixin/lib/promgrafonnet"

		resolveRepo, err := m.repoRoot(ctx, repo)

		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(resolveRepo).To(Equal("github.com/kubernetes-monitoring/kubernetes-mixin"))
	})

	t.Run("should get sub mod", func(t *testing.T) {
		mod, err := m.Get(ctx, "github.com/grafana/jsonnet-libs/grafana", "master", nil)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(mod.Module).To(Equal("github.com/grafana/jsonnet-libs/grafana"))
		NewWithT(t).Expect(mod.Repo).To(Equal("github.com/grafana/jsonnet-libs"))
	})

	t.Run("should get sub go mod", func(t *testing.T) {
		mod, err := m.Get(ctx, "github.com/prometheus/node_exporter/docs/node-mixin", "master", nil)
		NewWithT(t).Expect(err).To(BeNil())

		NewWithT(t).Expect(mod.Module).To(Equal("github.com/prometheus/node_exporter/docs/node-mixin"))
		NewWithT(t).Expect(mod.Repo).To(Equal("github.com/prometheus/node_exporter/docs/node-mixin"))
	})
}
