// Public CoreML inference interface. Wraps the vendored gomlx/go-coreml runtime
// so that NO go-coreml type appears in our exported API (SPEC §V16, T15).

//go:build coreml && darwin && arm64
// +build coreml,darwin,arm64

package coreml

import (
	"fmt"

	gcmodel "github.com/gomlx/go-coreml/model"
	gcruntime "github.com/gomlx/go-coreml/runtime"
)

// ComputeUnits selects which Apple compute units CoreML may dispatch to. CoreML
// still partitions the graph and may fall back; this is a hint, not a guarantee.
type ComputeUnits int

const (
	// ComputeAll uses ANE + GPU + CPU (default).
	ComputeAll ComputeUnits = iota
	// ComputeCPUOnly forces CPU (useful for debugging / reference).
	ComputeCPUOnly
	// ComputeCPUAndGPU uses CPU + GPU, skipping the Neural Engine.
	ComputeCPUAndGPU
	// ComputeCPUAndANE uses CPU + ANE, skipping the GPU.
	ComputeCPUAndANE
)

func (c ComputeUnits) toRuntime() (gcruntime.ComputeUnits, error) {
	switch c {
	case ComputeAll:
		return gcruntime.ComputeAll, nil
	case ComputeCPUOnly:
		return gcruntime.ComputeCPUOnly, nil
	case ComputeCPUAndGPU:
		return gcruntime.ComputeCPUAndGPU, nil
	case ComputeCPUAndANE:
		return gcruntime.ComputeCPUAndANE, nil
	default:
		return 0, fmt.Errorf("coreml: unknown ComputeUnits %d", int(c))
	}
}

// Model is a compiled CoreML model ready for inference. Construct it with Export
// (SPEC T16). It holds no exported go-coreml types.
type Model struct {
	rt   *gcruntime.Runtime
	exec *gcruntime.Executable
}

// Predict runs inference. Inputs and outputs are keyed by the model's input /
// output names; each value is a flat row-major float32 tensor.
func (m *Model) Predict(inputs map[string][]float32) (map[string][]float32, error) {
	if m == nil || m.exec == nil {
		return nil, fmt.Errorf("coreml: model closed")
	}
	feeds := make(map[string]any, len(inputs))
	for k, v := range inputs {
		feeds[k] = v
	}
	raw, err := m.exec.Run(feeds)
	if err != nil {
		return nil, fmt.Errorf("coreml: predict: %w", err)
	}
	out := make(map[string][]float32, len(raw))
	for k, v := range raw {
		f, ok := v.([]float32)
		if !ok {
			return nil, fmt.Errorf("coreml: output %q is %T, want []float32", k, v)
		}
		out[k] = f
	}
	return out, nil
}

// Close releases the compiled model + runtime. Idempotent.
func (m *Model) Close() error {
	if m == nil {
		return nil
	}
	var err error
	if m.exec != nil {
		err = m.exec.Close()
		m.exec = nil
	}
	if m.rt != nil {
		m.rt.Close()
		m.rt = nil
	}
	return err
}

// compileBuilder compiles a go-coreml MIL builder into a Model with the given
// compute units. Internal — the graph->MIL Export (T16) calls this; it keeps
// go-coreml types out of the public surface.
func compileBuilder(b *gcmodel.Builder, units ComputeUnits) (*Model, error) {
	ru, err := units.toRuntime()
	if err != nil {
		return nil, err
	}
	rt := gcruntime.New(gcruntime.WithComputeUnits(ru))
	exec, err := rt.Compile(b)
	if err != nil {
		rt.Close()
		return nil, fmt.Errorf("coreml: compile: %w", err)
	}
	return &Model{rt: rt, exec: exec}, nil
}
