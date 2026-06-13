//go:build metal && darwin && arm64
// +build metal,darwin,arm64

package metal

import (
	"math"
	"testing"

	"gorgonia.org/tensor"
)

// tensor ops on metal.Engine tensors dispatch elementwise Add/Sub/Mul to the GPU
// (plain same-shape float32 case) and match the CPU result (SPEC §V33).
func TestEngineElementwiseDispatch(t *testing.T) {
	e, err := NewEngine()
	if err != nil {
		t.Skipf("metal unavailable: %v", err)
	}
	defer e.Close()

	av := []float32{1, 2, 3, 4, 5, 6}
	bv := []float32{10, 20, 30, 40, 50, 60}
	mk := func(d []float32) *tensor.Dense {
		return tensor.New(tensor.WithEngine(e), tensor.WithShape(2, 3),
			tensor.WithBacking(append([]float32(nil), d...)))
	}

	cases := []struct {
		name string
		op   func(a, b tensor.Tensor) (tensor.Tensor, error)
		want func(x, y float32) float32
	}{
		{"Add", func(a, b tensor.Tensor) (tensor.Tensor, error) { return e.Add(a, b) }, func(x, y float32) float32 { return x + y }},
		{"Sub", func(a, b tensor.Tensor) (tensor.Tensor, error) { return e.Sub(a, b) }, func(x, y float32) float32 { return x - y }},
		{"Mul", func(a, b tensor.Tensor) (tensor.Tensor, error) { return e.Mul(a, b) }, func(x, y float32) float32 { return x * y }},
	}
	for _, c := range cases {
		before := e.ElemwiseOnGPU()
		out, err := c.op(mk(av), mk(bv))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if e.ElemwiseOnGPU()-before != 1 {
			t.Fatalf("%s did not dispatch to GPU", c.name)
		}
		got := out.Data().([]float32)
		for i := range av {
			if w := c.want(av[i], bv[i]); math.Abs(float64(got[i]-w)) > 1e-4 {
				t.Fatalf("%s idx %d: gpu=%v cpu=%v", c.name, i, got[i], w)
			}
		}
		t.Logf("%s on GPU OK: %v", c.name, got)
	}
}
