package dataviz

import (
	"context"
	"log"
	"os"

	"github.com/jxsl13/gorgonia"
	xvm "github.com/jxsl13/gorgonia/x/vm"
)

func ExampleDumpTrace() {
	g := gorgonia.NewGraph()
	// Add elements
	ctx, traceC := xvm.WithTracing(context.Background())
	defer xvm.CloseTracing(ctx)
	traces := make([]xvm.Trace, 0)
	go func() {
		for v := range traceC {
			traces = append(traces, v)
		}
	}()
	machine := xvm.NewMachine(g)
	err := machine.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
	machine.Close()
	DumpTrace(traces, g, os.Stdout)
}
