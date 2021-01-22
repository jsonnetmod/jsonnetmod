package tankax

import (
	"fmt"
	"os"
	"testing"

	"github.com/grafana/tanka/pkg/tanka"
	. "github.com/onsi/gomega"
)

func TestLoadComponent(t *testing.T) {
	_ = os.Chdir("../../")

	ctx := NewContext(ContextOpts{
		JsonnetHome: []string{
			"./lib",
			"./vendor",
		},
	})

	l, err := ctx.LoadClusterComponent("./jsonnet/clusters/demo/hello-world.jsonnet", tanka.Opts{})
	NewWithT(t).Expect(err).To(BeNil())
	fmt.Println(l.Resources().String())
}
