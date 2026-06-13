//go:build metal && darwin && arm64

package metal

import (
	"math"
	"testing"

	G "github.com/jxsl13/gorgonia"
	"gorgonia.org/tensor"
)

// A full gorgonia graph whose bound values use the Metal engine must run its
// MatMul on the GPU (no CUDA-style device-transfer machinery needed, because the
// engine is host-accessible) and match the CPU result. SPEC §V14, T23.
func TestGorgoniaGraphRunsMatMulOnGPU(t *testing.T) {
	e, err := NewEngine()
	if err != nil {
		t.Skipf("metal unavailable: %v", err)
	}
	defer e.Close()

	g := G.NewGraph()
	x := G.NewMatrix(g, tensor.Float32, G.WithShape(2, 3), G.WithName("x"))
	w := G.NewMatrix(g, tensor.Float32, G.WithShape(3, 2), G.WithName("w"))
	y, err := G.Mul(x, w)
	if err != nil {
		t.Fatalf("Mul: %v", err)
	}

	xData := []float32{1, 2, 3, 4, 5, 6}
	wData := []float32{1, 0, 0, 1, 1, 1}
	// Bind values backed by the Metal engine.
	xT := tensor.New(tensor.WithEngine(e), tensor.WithShape(2, 3), tensor.WithBacking(append([]float32(nil), xData...)))
	wT := tensor.New(tensor.WithEngine(e), tensor.WithShape(3, 2), tensor.WithBacking(append([]float32(nil), wData...)))
	G.Let(x, xT)
	G.Let(w, wT)

	before := e.MatMulsOnGPU()
	// The TapeMachine overrides value engines with its own (default
	// StandardEngine); pass the Metal engine so ops dispatch to the GPU.
	vm := G.NewTapeMachine(g, G.WithEngine(e))
	if err := vm.RunAll(); err != nil {
		vm.Close()
		t.Fatalf("RunAll: %v", err)
	}
	vm.Close()

	if got := e.MatMulsOnGPU() - before; got != 1 {
		t.Fatalf("expected 1 GPU matmul dispatch, got %d (graph did not run on Metal)", got)
	}

	out := y.Value().Data().([]float32)
	want := []float32{4, 5, 10, 11}
	for i := range want {
		if math.Abs(float64(out[i]-want[i])) > 1e-4 {
			t.Fatalf("idx %d: gpu=%v cpu=%v", i, out[i], want[i])
		}
	}
	t.Logf("gorgonia graph ran MatMul on GPU, result %v", out)
}
