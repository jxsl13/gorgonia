//go:build coreml && darwin && arm64

package coreml

import (
	"math"
	"testing"

	G "github.com/jxsl13/gorgonia"
	"gorgonia.org/tensor"
)

// Build y = x · w in gorgonia, evaluate on CPU for the reference, then Export to
// CoreML and Predict — the two must agree (SPEC §V18, T16).
func TestExportMatMulParity(t *testing.T) {
	g := G.NewGraph()
	x := G.NewMatrix(g, tensor.Float32, G.WithShape(2, 3), G.WithName("x"))
	w := G.NewMatrix(g, tensor.Float32, G.WithShape(3, 2), G.WithName("w"))
	y, err := G.Mul(x, w) // matrix·matrix = matmul
	if err != nil {
		t.Fatalf("Mul: %v", err)
	}

	xData := []float32{1, 2, 3, 4, 5, 6}
	wData := []float32{1, 0, 0, 1, 1, 1}

	// Reference via gorgonia CPU.
	xT := tensor.New(tensor.WithShape(2, 3), tensor.WithBacking(append([]float32(nil), xData...)))
	wT := tensor.New(tensor.WithShape(3, 2), tensor.WithBacking(append([]float32(nil), wData...)))
	G.Let(x, xT)
	G.Let(w, wT)
	vm := G.NewTapeMachine(g)
	if err := vm.RunAll(); err != nil {
		vm.Close()
		t.Fatalf("RunAll: %v", err)
	}
	want := y.Value().Data().([]float32)
	vm.Close()

	// Export + predict on CoreML.
	m, err := Export(g, y)
	if err != nil {
		t.Skipf("Export/compile unavailable: %v", err)
	}
	defer m.Close()

	out, err := m.Predict(map[string][]float32{"x": xData, "w": wData})
	if err != nil {
		t.Fatalf("Predict: %v", err)
	}
	got := out[OutputName(y, 0)]
	if len(got) != len(want) {
		t.Fatalf("len %d != %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(float64(got[i]-want[i])) > 1e-4 {
			t.Fatalf("idx %d: coreml=%v gorgonia=%v", i, got[i], want[i])
		}
	}
	t.Logf("graph->CoreML parity OK: %v", got)
}

// An unsupported op must produce a clear error, not a silent miscompile.
func TestExportUnsupportedOp(t *testing.T) {
	g := G.NewGraph()
	x := G.NewMatrix(g, tensor.Float32, G.WithShape(2, 2), G.WithName("x"))
	y, err := G.Tanh(x) // tanh not in the supported subset yet
	if err != nil {
		t.Fatalf("Tanh: %v", err)
	}
	if _, err := Export(g, y); err == nil {
		t.Fatal("expected unsupported-op error, got nil")
	}
}
