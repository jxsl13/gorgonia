// Example: export a gorgonia graph to CoreML and run inference on Apple Silicon
// (Neural Engine + GPU + CPU). Build/run with the `coreml` tag on darwin/arm64:
//
//	go run -tags coreml ./examples/coreml
//
// Needs only the Command Line Tools (runtime compile), not full Xcode.
// SPEC: T17.

//go:build coreml && darwin && arm64
// +build coreml,darwin,arm64

package main

import (
	"fmt"
	"log"

	"github.com/jxsl13/gorgonia/coreml"

	G "github.com/jxsl13/gorgonia"
	"gorgonia.org/tensor"
)

func main() {
	// Build y = x · w in gorgonia.
	g := G.NewGraph()
	x := G.NewMatrix(g, tensor.Float32, G.WithShape(2, 3), G.WithName("x"))
	w := G.NewMatrix(g, tensor.Float32, G.WithShape(3, 2), G.WithName("w"))
	y, err := G.Mul(x, w)
	if err != nil {
		log.Fatal(err)
	}

	// Export to CoreML (default compute units = ANE + GPU + CPU).
	m, err := coreml.Export(g, y)
	if err != nil {
		log.Fatalf("export: %v", err)
	}
	defer m.Close()

	out, err := m.Predict(map[string][]float32{
		"x": {1, 2, 3, 4, 5, 6},
		"w": {1, 0, 0, 1, 1, 1},
	})
	if err != nil {
		log.Fatalf("predict: %v", err)
	}
	fmt.Println("y = x·w on CoreML:", out[coreml.OutputName(y, 0)]) // [4 5 10 11]
}
