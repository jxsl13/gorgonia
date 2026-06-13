// Graph -> CoreML MIL translator. Walks a gorgonia ExprGraph and emits an
// equivalent MIL program, compiled into a Model. Supported op subset:
// add/sub/mul/div + matmul (SPEC T16); unsupported ops -> clear error.

//go:build coreml && darwin && arm64
// +build coreml,darwin,arm64

package coreml

import (
	"fmt"
	"strings"

	G "github.com/jxsl13/gorgonia"
	gcmodel "github.com/gomlx/go-coreml/model"
	"gorgonia.org/tensor"
)

// milName sanitizes s into a valid MIL identifier: [A-Za-z_][A-Za-z0-9_]*.
// gorgonia auto-names nodes with math symbols (e.g. "√", "×") that the MIL
// parser rejects.
func milName(s string) string {
	var b strings.Builder
	for i, r := range s {
		ok := r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "_"
	}
	return b.String()
}

// Export translates the sub-graph producing outputs into a CoreML model using
// all available compute units (ANE+GPU+CPU). Input/constant nodes must be
// float32; every op must be supported (see OpKind) or Export errors.
func Export(g *G.ExprGraph, outputs ...*G.Node) (*Model, error) {
	return ExportWithComputeUnits(ComputeAll, g, outputs...)
}

// ExportWithComputeUnits is Export with an explicit compute-unit hint.
func ExportWithComputeUnits(units ComputeUnits, g *G.ExprGraph, outputs ...*G.Node) (*Model, error) {
	if len(outputs) == 0 {
		return nil, fmt.Errorf("coreml: Export needs at least one output")
	}
	b := gcmodel.NewBuilder("main")
	vals := make(map[*G.Node]*gcmodel.Value)

	var visit func(n *G.Node) error
	visit = func(n *G.Node) error {
		if _, done := vals[n]; done {
			return nil
		}
		ins := n.InputNodes()
		for _, c := range ins {
			if err := visit(c); err != nil {
				return err
			}
		}
		dt, err := toDType(n.Dtype())
		if err != nil {
			return fmt.Errorf("coreml: node %q: %w", n.Name(), err)
		}
		switch {
		case n.IsVar():
			name := n.Name()
			if name == "" {
				return fmt.Errorf("coreml: input node id %d has no name (needed as Predict key)", n.ID())
			}
			vals[n] = b.Input(milName(name), dt, shape64(n.Shape())...)
		case n.Op() == nil: // constant
			data, err := float32Data(n)
			if err != nil {
				return fmt.Errorf("coreml: const node %q: %w", n.Name(), err)
			}
			vals[n] = b.Const(constName(n), dt, shape64(n.Shape()), data)
		default:
			kind, ok := G.OpKind(n.Op())
			if !ok {
				return fmt.Errorf("coreml: unsupported op %q (%T) at node %q", n.Op(), n.Op(), n.Name())
			}
			x := vals[ins[0]]
			var y *gcmodel.Value
			if len(ins) > 1 {
				y = vals[ins[1]]
			}
			switch kind {
			case "add":
				vals[n] = b.Add(x, y)
			case "sub":
				vals[n] = b.Sub(x, y)
			case "mul":
				vals[n] = b.Mul(x, y)
			case "div":
				vals[n] = b.Div(x, y)
			case "matmul":
				vals[n] = b.MatMul(x, y)
			default:
				return fmt.Errorf("coreml: op kind %q recognised but not wired", kind)
			}
		}
		return nil
	}

	for _, o := range outputs {
		if err := visit(o); err != nil {
			return nil, err
		}
	}
	for i, o := range outputs {
		b.Output(outputName(o, i), vals[o])
	}
	return compileBuilder(b, units)
}

// OutputName returns the Predict-result key for the i-th Export output.
func OutputName(o *G.Node, i int) string { return outputName(o, i) }

func outputName(o *G.Node, i int) string {
	if n := o.Name(); n != "" {
		return milName(n)
	}
	return fmt.Sprintf("output%d", i)
}

func constName(n *G.Node) string {
	if nm := n.Name(); nm != "" {
		return milName(nm)
	}
	return fmt.Sprintf("const%d", n.ID())
}

func toDType(dt tensor.Dtype) (gcmodel.DType, error) {
	switch dt {
	case tensor.Float32:
		return gcmodel.Float32, nil
	default:
		return 0, fmt.Errorf("unsupported dtype %v (only Float32)", dt)
	}
}

func shape64(s tensor.Shape) []int64 {
	out := make([]int64, len(s))
	for i, d := range s {
		out[i] = int64(d)
	}
	return out
}

func float32Data(n *G.Node) ([]float32, error) {
	v := n.Value()
	if v == nil {
		return nil, fmt.Errorf("node has no value")
	}
	switch d := v.Data().(type) {
	case []float32:
		return d, nil
	case float32:
		return []float32{d}, nil
	default:
		return nil, fmt.Errorf("value data is %T, want float32", v.Data())
	}
}
