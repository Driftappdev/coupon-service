package tracing

import (
	"io"
	"os"

	gotracing "github.com/driftappdev/libpackage/gotracing"
)

func NewProvider(service string) *gotracing.Provider {
	w := io.Writer(os.Stdout)
	return gotracing.NewProvider(gotracing.ProviderConfig{
		Sampler: gotracing.AlwaysSample{},
		Exporters: []gotracing.Exporter{
			&gotracing.StdoutExporter{W: w},
		},
	})
}
