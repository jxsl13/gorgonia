package gorgonia

// Introspection helpers used by backends that translate an ExprGraph to another
// representation (e.g. the CoreML export in package coreml). They expose, via
// the public API, what a separate package cannot reach: a node's ordered inputs
// and a stable op-kind name. SPEC T16.

// InputNodes returns the ordered input nodes (children) of n. For an input or
// constant node it returns nil.
func (n *Node) InputNodes() Nodes { return n.children }

// OpKind returns a stable, backend-neutral name for op ("add", "sub", "mul",
// "div", "matmul", ...). ok is false for ops a translator does not (yet)
// recognise, so callers can emit a clear "unsupported op" error.
func OpKind(op Op) (kind string, ok bool) {
	switch o := op.(type) {
	case elemBinOp:
		switch o.binOpType() {
		case addOpType:
			return "add", true
		case subOpType:
			return "sub", true
		case mulOpType:
			return "mul", true
		case divOpType:
			return "div", true
		}
	case linAlgBinOp:
		switch o.āBinaryOperator {
		case matMulOperator:
			return "matmul", true
		}
	}
	return "", false
}
