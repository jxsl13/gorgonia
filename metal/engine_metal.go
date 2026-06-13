// Metal-backed tensor.Engine. Embeds tensor.StdEng for host memory management
// and overrides the heavy ops (MatMul) to run on the GPU, so tensors created
// WithEngine(metal.Engine) auto-dispatch those ops to Metal. SPEC T22, §V14.
//
// This is the op-level Engine. Full gorgonia TapeMachine device-transfer wiring
// (mirroring cuda/*_cuda.go) builds on top of it and is a separate, larger step.

//go:build metal && darwin && arm64

package metal

import (
	"fmt"

	"gorgonia.org/tensor"
)

// Engine is a tensor.Engine that runs supported ops on the Apple GPU. Memory is
// host-accessible (inherited from tensor.StdEng); compute ops copy to the GPU,
// run, and copy back. Not safe for concurrent use.
type Engine struct {
	tensor.StdEng
	dev         *Device
	matmulOnGPU int // count of MatMul ops dispatched to the GPU (for tests/metrics)
	elemOnGPU   int // count of elementwise ops dispatched to the GPU
}

// MatMulsOnGPU reports how many MatMul ops this engine ran on the GPU. Lets
// callers/tests confirm a gorgonia graph actually dispatched to Metal.
func (e *Engine) MatMulsOnGPU() int { return e.matmulOnGPU }

// ElemwiseOnGPU reports how many elementwise ops ran on the GPU.
func (e *Engine) ElemwiseOnGPU() int { return e.elemOnGPU }

// elementwise runs a + b (op 0), a - b (1) or a * b (2) on the GPU for the
// plain same-shape float32 case (no FuncOpts). ok=false means "not GPU-eligible,
// caller should delegate to the embedded StdEng on the CPU".
//
// NOTE: buffers are copied to the GPU and back per call, so for typical sizes
// the copy dominates and this is not faster than the CPU. A real speedup needs
// persistent GPU-resident tensors (future work); this provides the dispatch
// path + correctness (SPEC §V33).
func (e *Engine) elementwise(op int, a, b tensor.Tensor, opts []tensor.FuncOpt) (tensor.Tensor, bool, error) {
	if e.dev == nil || len(opts) != 0 {
		return nil, false, nil
	}
	ad, ok1 := a.Data().([]float32)
	bd, ok2 := b.Data().([]float32)
	if !ok1 || !ok2 || !a.Shape().Eq(b.Shape()) {
		return nil, false, nil
	}
	out, err := e.dev.run(op, ad, bd)
	if err != nil {
		return nil, false, err
	}
	e.elemOnGPU++
	res := tensor.New(tensor.WithShape(a.Shape()...), tensor.WithBacking(out), tensor.WithEngine(e))
	return res, true, nil
}

// Add implements tensor.Adder: GPU for the plain same-shape float32 case, else
// the embedded StdEng (CPU) handles scalars, broadcasting and reuse/incr opts.
func (e *Engine) Add(a, b tensor.Tensor, opts ...tensor.FuncOpt) (tensor.Tensor, error) {
	if r, ok, err := e.elementwise(0, a, b, opts); ok || err != nil {
		return r, err
	}
	return e.StdEng.Add(a, b, opts...)
}

// Sub implements tensor.Suber (GPU plain case, else StdEng).
func (e *Engine) Sub(a, b tensor.Tensor, opts ...tensor.FuncOpt) (tensor.Tensor, error) {
	if r, ok, err := e.elementwise(1, a, b, opts); ok || err != nil {
		return r, err
	}
	return e.StdEng.Sub(a, b, opts...)
}

// Mul implements tensor.Multiplier (GPU plain case, else StdEng).
func (e *Engine) Mul(a, b tensor.Tensor, opts ...tensor.FuncOpt) (tensor.Tensor, error) {
	if r, ok, err := e.elementwise(2, a, b, opts); ok || err != nil {
		return r, err
	}
	return e.StdEng.Mul(a, b, opts...)
}

// NewEngine creates a Metal-backed tensor engine.
func NewEngine() (*Engine, error) {
	d, err := New()
	if err != nil {
		return nil, err
	}
	return &Engine{dev: d}, nil
}

// Device returns the underlying GPU device.
func (e *Engine) Device() *Device { return e.dev }

// Close releases the GPU device.
func (e *Engine) Close() {
	if e.dev != nil {
		e.dev.Close()
		e.dev = nil
	}
}

// MatMul implements tensor.MatMuler: prealloc = a · b, computed on the GPU.
// a is m×k, b is k×n, prealloc is m×n. float32 only.
func (e *Engine) MatMul(a, b, prealloc tensor.Tensor) error {
	if e.dev == nil {
		return fmt.Errorf("metal engine: closed")
	}
	ad, ok := a.Data().([]float32)
	if !ok {
		return fmt.Errorf("metal engine: MatMul needs float32, got %T", a.Data())
	}
	bd, ok := b.Data().([]float32)
	if !ok {
		return fmt.Errorf("metal engine: MatMul needs float32, got %T", b.Data())
	}
	pd, ok := prealloc.Data().([]float32)
	if !ok {
		return fmt.Errorf("metal engine: MatMul prealloc needs float32, got %T", prealloc.Data())
	}
	as, bs := a.Shape(), b.Shape()
	if len(as) != 2 || len(bs) != 2 {
		return fmt.Errorf("metal engine: MatMul needs 2-D, got %v · %v", as, bs)
	}
	m, k, n := as[0], as[1], bs[1]
	out, err := e.dev.MatMul(ad, bd, m, n, k)
	if err != nil {
		return err
	}
	copy(pd, out)
	e.matmulOnGPU++
	return nil
}
