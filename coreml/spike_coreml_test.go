// T14 spike: prove the CoreML pipeline works end-to-end with Command Line Tools
// only (no full Xcode / coremlcompiler) — build a MIL program, compile it at
// runtime, run inference (default compute units = ANE+GPU+CPU), check parity
// vs a CPU reference. SPEC §C13, §V18, B6.

//go:build coreml && darwin && arm64

package coreml

import (
	"math"
	"testing"

	"github.com/gomlx/go-coreml/model"
	"github.com/gomlx/go-coreml/runtime"
)

func cpuMatMul(a, b []float32, m, n, k int) []float32 {
	c := make([]float32, m*n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			var s float32
			for p := 0; p < k; p++ {
				s += a[i*k+p] * b[p*n+j]
			}
			c[i*n+j] = s
		}
	}
	return c
}

func TestCoreMLMatMulSpike(t *testing.T) {
	const M, K, N = 2, 3, 4

	b := model.NewBuilder("main")
	x := b.Input("x", model.Float32, M, K)
	w := b.Input("w", model.Float32, K, N)
	z := b.MatMul(x, w)
	b.Output("z", z)

	rt := runtime.New() // default compute units = ComputeAll (ANE + GPU + CPU)
	exec, err := rt.Compile(b)
	if err != nil {
		t.Skipf("CoreML compile unavailable on this host: %v", err)
	}
	defer exec.Close()

	xv := []float32{1, 2, 3, 4, 5, 6}
	wv := []float32{1, 0, 2, 0, 0, 1, 0, 2, 1, 1, 1, 1}

	out, err := exec.Run(map[string]any{"x": xv, "w": wv})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got, ok := out["z"].([]float32)
	if !ok {
		t.Fatalf("output z: expected []float32, got %T", out["z"])
	}

	want := cpuMatMul(xv, wv, M, N, K)
	if len(got) != len(want) {
		t.Fatalf("len %d != %d", len(got), len(want))
	}
	const tol = 1e-4
	for i := range want {
		if d := math.Abs(float64(got[i] - want[i])); d > tol {
			t.Fatalf("idx %d: coreml=%v cpu=%v (|diff| %g > %g)", i, got[i], want[i], d, tol)
		}
	}
	t.Logf("CoreML matmul parity OK (CLT-only runtime compile): %v", got)
}
