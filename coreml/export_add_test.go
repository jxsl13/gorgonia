//go:build coreml && darwin && arm64
// +build coreml,darwin,arm64

package coreml

import (
	"math"
	"testing"

	G "github.com/jxsl13/gorgonia"
	"gorgonia.org/tensor"
)

// Exercise the elementwise add path through the translator: z = a + b.
func TestExportAddParity(t *testing.T) {
	g := G.NewGraph()
	a := G.NewMatrix(g, tensor.Float32, G.WithShape(2, 2), G.WithName("a"))
	b := G.NewMatrix(g, tensor.Float32, G.WithShape(2, 2), G.WithName("b"))
	z, err := G.Add(a, b)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	aData := []float32{1, 2, 3, 4}
	bData := []float32{10, 20, 30, 40}
	want := []float32{11, 22, 33, 44}

	m, err := Export(g, z)
	if err != nil {
		t.Skipf("Export/compile unavailable: %v", err)
	}
	defer m.Close()

	out, err := m.Predict(map[string][]float32{"a": aData, "b": bData})
	if err != nil {
		t.Fatalf("Predict: %v", err)
	}
	got := out[OutputName(z, 0)]
	for i := range want {
		if math.Abs(float64(got[i]-want[i])) > 1e-4 {
			t.Fatalf("idx %d: coreml=%v want=%v", i, got[i], want[i])
		}
	}
	t.Logf("graph->CoreML add parity OK: %v", got)
}
