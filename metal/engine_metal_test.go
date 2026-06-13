//go:build metal && darwin && arm64
// +build metal,darwin,arm64

package metal

import (
	"math"
	"testing"

	"gorgonia.org/tensor"
)

// Tensors created WithEngine(metal.Engine) must dispatch MatMul to the GPU and
// agree with the CPU result (SPEC §V14, T22).
func TestEngineMatMulDispatch(t *testing.T) {
	e, err := NewEngine()
	if err != nil {
		t.Skipf("metal unavailable: %v", err)
	}
	defer e.Close()

	aData := []float32{1, 2, 3, 4, 5, 6}    // 2x3
	bData := []float32{1, 0, 0, 1, 1, 1}    // 3x2
	a := tensor.New(tensor.WithEngine(e), tensor.WithShape(2, 3), tensor.WithBacking(append([]float32(nil), aData...)))
	b := tensor.New(tensor.WithEngine(e), tensor.WithShape(3, 2), tensor.WithBacking(append([]float32(nil), bData...)))

	c, err := a.MatMul(b) // dispatches to Engine.MatMul on the GPU
	if err != nil {
		t.Fatalf("MatMul: %v", err)
	}

	got := c.Data().([]float32)
	// CPU reference
	want := make([]float32, 2*2)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			var s float32
			for p := 0; p < 3; p++ {
				s += aData[i*3+p] * bData[p*2+j]
			}
			want[i*2+j] = s
		}
	}
	for i := range want {
		if math.Abs(float64(got[i]-want[i])) > 1e-4 {
			t.Fatalf("idx %d: gpu-engine=%v cpu=%v", i, got[i], want[i])
		}
	}
	t.Logf("tensor.Engine GPU matmul dispatch OK: %v", got)
}
