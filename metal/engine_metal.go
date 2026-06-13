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
}

// MatMulsOnGPU reports how many MatMul ops this engine ran on the GPU. Lets
// callers/tests confirm a gorgonia graph actually dispatched to Metal.
func (e *Engine) MatMulsOnGPU() int { return e.matmulOnGPU }

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
