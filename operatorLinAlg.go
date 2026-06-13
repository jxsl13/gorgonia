package gorgonia

import (
	"fmt"

	"github.com/chewxy/hm"
	"gorgonia.org/tensor"
)

// ā and Ā are used to denote that it's a matrix/vector type.
// if you want to type it, it's Latin Letter A with Macron (lowercase and capital)
// Codepoints : U+101 for the small one, and U+100 for the capital one

type āBinaryOperator byte

const (
	matMulOperator        āBinaryOperator = iota // emits S/DGEMM BLAS calls
	matVecMulOperator                            // emits S/DGEMV BLAS calls
	vecDotOperator                               // emits S/DDOT BLAS calls
	outerProdOperator                            // emits S/DGER BLAS calls
	batchedMatMulOperator                        // just S/GEMM BLAS calls in a loop

	maxĀBinaryOperator // delimits all possible linalg operators. Add above this line
)

func (op āBinaryOperator) String() string {
	if op >= maxĀBinaryOperator {
		return "UNSUPPORTED LINEAR ALGEBRA OPERATOR"
	}
	return āBinOpStrs[op]
}

func (op āBinaryOperator) Type() hm.Type {
	if op >= maxĀBinaryOperator {
		panic("UNSUPPORTED LINEAR ALGEBRA OPERATOR")
	}
	return āBinOpTypes[op]()
}

func (op āBinaryOperator) DiffWRT(inputs int) []bool {
	if inputs != 2 {
		panic("binary linear algebra operator only supports two and only two inputs")
	}

	if op >= maxĀBinaryOperator {
		panic("Unsupported unary operator is not differentiable")
	}
	return []bool{true, true}
}

// todo: write explanation.
func matMulDiffExpr(transA, transB bool, x, y, z, gradZ *Node) (retVal Nodes, err error) {
	var dzdx, dzdy *Node
	op := linAlgBinOp{
		āBinaryOperator: matMulOperator,
	}

	switch {
	case transA && transB:
		op.transA = transA
		op.transB = transB
		if dzdx, err = binOpNode(op, y, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
		if dzdy, err = binOpNode(op, gradZ, x); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	case !transA && transB:
		if dzdx, err = binOpNode(op, gradZ, y); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}

		op.transA = true
		if dzdy, err = binOpNode(op, gradZ, x); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	case transA && !transB:
		op.transB = true
		if dzdx, err = binOpNode(op, y, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}

		op.transB = false
		if dzdy, err = binOpNode(op, x, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	case !transA && !transB:
		// dzdy
		op.transA = false
		op.transB = true
		if dzdx, err = binOpNode(op, gradZ, y); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
		// do dzdx
		op.transA = true
		op.transB = false
		if dzdy, err = binOpNode(op, x, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	}
	retVal = Nodes{dzdx, dzdy}
	return
}

func matMulDiff(ctx ExecutionContext, transA, transB bool, x, y, z *Node) (err error) {
	xdv, ydv, zdv := getDV3(x, y, z)

	op := linAlgBinOp{
		āBinaryOperator: matMulOperator,
	}

	switch {
	case transA && transB:
		op.transA = transA
		op.transB = transB

		// dzdx
		err = op.IncrDo(xdv.d, ydv.Value, zdv.d)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		// dzdy
		err = op.IncrDo(ydv.d, zdv.d, xdv.Value)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", y, err)
		}

		return

	case !transA && transB:
		// dzdx
		err = op.IncrDo(xdv.d, zdv.d, ydv.Value)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		// dzdy
		op.transA = true
		err = op.IncrDo(ydv.d, zdv.d, xdv.Value)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		return

	case transA && !transB:
		// dzdx
		op.transB = true
		err = op.IncrDo(xdv.d, ydv.Value, zdv.d)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		// dzdy
		op.transA = false
		op.transB = false
		err = op.IncrDo(ydv.d, xdv.Value, zdv.d)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}
		return
	case !transA && !transB:
		op.transB = true
		err = op.IncrDo(xdv.d, zdv.d, ydv.Value)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		op.transA = true
		op.transB = false
		err = op.IncrDo(ydv.d, xdv.Value, zdv.d)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}
		return
	}

	panic("unreachable")
}

func matVecMulDiffExpr(transA, transB bool, x, y, z, gradZ *Node) (retVal Nodes, err error) {
	var dzdx, dzdy *Node
	if transA {
		dzdx, err = OuterProd(y, gradZ)
	} else {
		dzdx, err = OuterProd(gradZ, y)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", "Failed to carry outper product", err)
	}

	op := linAlgBinOp{
		āBinaryOperator: matVecMulOperator,
		transA:          !transA,
	}

	if dzdy, err = binOpNode(op, x, gradZ); err != nil {
		return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
	}
	return Nodes{dzdx, dzdy}, nil
}

func matVecMulDiff(ctx ExecutionContext, transA, transB bool, x, y, z *Node) (err error) {
	xdv, ydv, zdv := getDV3(x, y, z)

	op := linAlgBinOp{
		āBinaryOperator: outerProdOperator,
	}

	if transA {
		err = op.IncrDo(xdv.d, ydv.Value, zdv.d)
	} else {
		err = op.IncrDo(xdv.d, zdv.d, ydv.Value)
	}
	if err = checkErrSetDeriv(err, xdv); err != nil {
		return fmt.Errorf(autodiffFail+": %w", x, err)
	}

	op = linAlgBinOp{
		āBinaryOperator: matVecMulOperator,
		transA:          !transA,
	}

	err = op.IncrDo(ydv.d, xdv.Value, zdv.d)
	if err = checkErrSetDeriv(err, ydv); err != nil {
		return fmt.Errorf(autodiffFail+": %w", x, err)
	}
	return
}

func vecDotDiffExpr(transA, transB bool, x, y, z, gradZ *Node) (retVal Nodes, err error) {
	var dzdx, dzdy *Node
	if dzdx, err = HadamardProd(y, gradZ); err == nil {
		if dzdy, err = HadamardProd(x, gradZ); err == nil {
			retVal = Nodes{dzdx, dzdy}
		} else {
			return nil, fmt.Errorf("%s: %w", "Failed to carry HadamardProd()", err)
		}
	} else {
		return nil, fmt.Errorf("%s: %w", "Failed to carry HadamardProd()", err)
	}
	return
}

func vecDotDiff(ctx ExecutionContext, transA, transB bool, x, y, z *Node) (err error) {
	xdv, ydv, zdv := getDV3(x, y, z)

	mul := newElemBinOp(mulOpType, x, z)
	err = mul.IncrDo(xdv.d, ydv.Value, zdv.d)
	if err = checkErrSetDeriv(err, xdv); err != nil {
		return fmt.Errorf(autodiffFail+": %w", x, err)
	}

	err = mul.IncrDo(ydv.d, xdv.Value, zdv.d)
	if err = checkErrSetDeriv(err, ydv); err != nil {
		return fmt.Errorf(autodiffFail+": %w", x, err)
	}
	return
}

func outerProdDiffExpr(transA, transB bool, x, y, z, gradZ *Node) (retVal Nodes, err error) {
	var dzdx, dzdy *Node
	if dzdx, err = Mul(x, gradZ); err == nil {
		if dzdy, err = Mul(y, gradZ); err == nil {
			retVal = Nodes{dzdx, dzdy}
		} else {
			return nil, fmt.Errorf("%s: %w", "Failed to carry Mul()", err)
		}
	} else {
		return nil, fmt.Errorf("%s: %w", "Failed to carry Mul()", err)
	}
	return
}

func outerProdDiff(ctx ExecutionContext, transA, transB bool, x, y, z *Node) (err error) {
	xdv, ydv, zdv := getDV3(x, y, z)

	mul := newElemBinOp(mulOpType, x, z)
	err = mul.IncrDo(xdv.d, xdv.Value, zdv.d)
	err = mul.IncrDo(xdv.d, ydv.Value, zdv.d)
	if err = checkErrSetDeriv(err, xdv); err != nil {
		return fmt.Errorf(autodiffFail+": %w", x, err)
	}

	err = mul.IncrDo(ydv.d, ydv.Value, zdv.d)
	if err = checkErrSetDeriv(err, ydv); err != nil {
		return fmt.Errorf(autodiffFail+": %w", x, err)
	}
	return
}

func batchedMatMulDiffExpr(transA, transB bool, x, y, z, gradZ *Node) (retVal Nodes, err error) {
	var dzdx, dzdy *Node
	op := linAlgBinOp{
		āBinaryOperator: batchedMatMulOperator,
	}

	switch {
	case transA && transB:
		op.transA = transA
		op.transB = transB
		if dzdx, err = binOpNode(op, y, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
		if dzdy, err = binOpNode(op, gradZ, x); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	case !transA && transB:
		if dzdx, err = binOpNode(op, gradZ, y); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}

		op.transA = true
		if dzdy, err = binOpNode(op, gradZ, x); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	case transA && !transB:
		op.transB = true
		if dzdx, err = binOpNode(op, y, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}

		op.transB = false
		if dzdy, err = binOpNode(op, x, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	case !transA && !transB:
		// dzdy
		op.transA = false
		op.transB = true
		if dzdx, err = binOpNode(op, gradZ, y); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
		// do dzdx
		op.transA = true
		op.transB = false
		if dzdy, err = binOpNode(op, x, gradZ); err != nil {
			return nil, fmt.Errorf(binOpNodeFail+": %w", op, err)
		}
	}
	retVal = Nodes{dzdx, dzdy}
	return
}

func batchedMatMulDiff(ctx ExecutionContext, transA, transB bool, x, y, z *Node) (err error) {
	xdv, ydv, zdv := getDV3(x, y, z)

	op := linAlgBinOp{
		āBinaryOperator: batchedMatMulOperator,
	}

	switch {
	case transA && transB:
		op.transA = transA
		op.transB = transB

		// dzdx
		err = op.IncrDo(xdv.d, ydv.Value, zdv.d)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		// dzdy
		err = op.IncrDo(ydv.d, zdv.d, xdv.Value)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", y, err)
		}

		return

	case !transA && transB:
		// dzdx
		err = op.IncrDo(xdv.d, zdv.d, ydv.Value)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		// dzdy
		op.transA = true
		err = op.IncrDo(ydv.d, zdv.d, xdv.Value)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		return

	case transA && !transB:
		// dzdx
		op.transB = true
		err = op.IncrDo(xdv.d, ydv.Value, zdv.d)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		// dzdy
		op.transA = false
		op.transB = false
		err = op.IncrDo(ydv.d, xdv.Value, zdv.d)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}
		return
	case !transA && !transB:
		op.transB = true
		err = op.IncrDo(xdv.d, zdv.d, ydv.Value)
		if err = checkErrSetDeriv(err, xdv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}

		op.transA = true
		op.transB = false
		err = op.IncrDo(ydv.d, xdv.Value, zdv.d)
		if err = checkErrSetDeriv(err, ydv); err != nil {
			return fmt.Errorf(autodiffFail+": %w", x, err)
		}
		return
	}

	panic("unreachable")
}

func reshape(name string, t tensor.Tensor, shape ...int) error {
	if t.Shape().Eq(shape) {
		return nil
	}
	if t.DataOrder().IsContiguous() {
		if err := t.Reshape(shape...); err != nil {
			return fmt.Errorf("Reshaping slice for %s failed: %w", name, err)
		}
	}
	return nil
}

func batchedMatMul(a, b, c tensor.Tensor, transA, transB, incr bool) (retVal tensor.Tensor, err error) {
	shapeA := a.Shape().Clone()
	shapeB := b.Shape().Clone()
	outer := shapeA[:len(shapeA)-2]
	innerA := shapeA[len(shapeA)-2:]
	innerB := shapeB[len(shapeB)-2:]
	if transA {
		innerA[0], innerA[1] = innerA[1], innerA[0]
	}
	if transB {
		innerB[0], innerB[1] = innerB[1], innerB[0]
	}

	if c == nil {
		newShape := append(outer.Clone(), innerA[0], innerB[1])
		c = tensor.New(tensor.Of(a.Dtype()), tensor.WithShape(newShape...), tensor.WithEngine(a.Engine()))
	}

	slices := make([]sli, len(outer))
	ss := make([]tensor.Slice, len(slices))
	for i := range slices {
		slices[i].end = slices[i].start + 1
		ss[i] = &slices[i]
	}

	var as, bs, cs tensor.Tensor
	for halt := false; !halt; halt = incrSlices(slices, outer) {
		if as, err = a.Slice(ss...); err != nil {
			return nil, fmt.Errorf("Slicing %v from a failed: %w", ss, err)
		}
		if bs, err = b.Slice(ss...); err != nil {
			return nil, fmt.Errorf("Slicing %v from b failed: %w", ss, err)
		}
		if cs, err = c.Slice(ss...); err != nil {
			return nil, fmt.Errorf("Slicing %v from c failed: %w", ss, err)
		}

		if transA {
			as.T()
		}
		if transB {
			bs.T()
		}

		// Reshape the result matrix slice in case matrices like 1x1 will be converted to scalar which results in
		// not satisfying matrix multiplication dimension requirements.
		if err := reshape("a", as, innerA...); err != nil {
			return nil, err
		}
		if err := reshape("b", bs, innerB...); err != nil {
			return nil, err
		}

		var fo tensor.FuncOpt
		if incr {
			fo = tensor.WithIncr(cs)
		} else {
			fo = tensor.WithReuse(cs)
		}

		if _, err = tensor.MatMul(as, bs, fo); err != nil {
			return nil, fmt.Errorf("MatMul on batch %v failed.: %w", ss, err)
		}

	}

	return c, nil
}

// incrSlices increments the slices. If everything has matched then return true
func incrSlices(a []sli, shp tensor.Shape) (halt bool) {
	for i := len(a) - 1; i >= 0; i-- {
		if shp[i]-a[i].start == 1 {
			a[i].start = 0
			a[i].end = 1
			if i == 0 {
				return true
			}
			continue
		}

		a[i].start++
		a[i].end = a[i].start + 1
		return false
	}
	return true
}
