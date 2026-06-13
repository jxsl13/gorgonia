//go:build coreml && darwin && arm64

package coreml

import (
	"math"
	"testing"

	gcmodel "github.com/gomlx/go-coreml/model"
)

// matmulBuilder builds x[2,3] · w[3,2] -> z[2,2].
func matmulBuilder() *gcmodel.Builder {
	b := gcmodel.NewBuilder("main")
	x := b.Input("x", gcmodel.Float32, 2, 3)
	w := b.Input("w", gcmodel.Float32, 3, 2)
	b.Output("z", b.MatMul(x, w))
	return b
}

func TestModelPredict_ComputeUnits(t *testing.T) {
	xv := []float32{1, 2, 3, 4, 5, 6}
	wv := []float32{1, 0, 0, 1, 1, 1}
	// z = x·w: row0 [1+3, 2+3]=[4,5]; row1 [4+6,5+6]=[10,11]
	want := []float32{4, 5, 10, 11}

	for _, units := range []ComputeUnits{ComputeAll, ComputeCPUOnly, ComputeCPUAndGPU} {
		m, err := compileBuilder(matmulBuilder(), units)
		if err != nil {
			t.Skipf("CoreML compile unavailable (units=%d): %v", units, err)
		}
		out, err := m.Predict(map[string][]float32{"x": xv, "w": wv})
		if err != nil {
			m.Close()
			t.Fatalf("units=%d predict: %v", units, err)
		}
		got := out["z"]
		if len(got) != len(want) {
			m.Close()
			t.Fatalf("units=%d: len %d != %d", units, len(got), len(want))
		}
		for i := range want {
			if math.Abs(float64(got[i]-want[i])) > 1e-4 {
				m.Close()
				t.Fatalf("units=%d idx %d: got %v want %v", units, i, got[i], want[i])
			}
		}
		m.Close()
		t.Logf("units=%d parity OK: %v", units, got)
	}
}
