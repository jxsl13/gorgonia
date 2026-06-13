package gorgonia

import (
	"fmt"

	"gorgonia.org/tensor"
)

// this file provides utility functions for solvers

func doL1Reg(w, g tensor.Tensor, l1reg any) (err error) {
	var l1regs tensor.Tensor
	if l1regs, err = tensor.Sign(w); err != nil {
		err = fmt.Errorf("%s: %w", signFail, err)
	}
	if _, err = tensor.Mul(l1reg, l1regs, tensor.WithIncr(g)); err != nil {
		return fmt.Errorf("%s: %w", pointWiseMulFail, err)
	}
	defer returnTensor(l1regs)
	return nil
}

func doL2Reg(w, g tensor.Tensor, l2reg any) (err error) {
	if _, err = tensor.Mul(w, l2reg, tensor.WithIncr(g)); err != nil {
		return fmt.Errorf("%s: %w", pointWiseMulFail, err)
	}
	return nil
}

func computeRecip(x float64, as tensor.Dtype) (retVal any, err error) {
	switch as {
	case tensor.Float64:
		return 1.0 / x, nil
	case tensor.Float32:
		return float32(1.0) / float32(x), nil
	default:
		return 0.0, fmt.Errorf("Unhandled Dtype %v for computeRecip", as)
	}
}

func divBatch(g tensor.Tensor, batch float64) (err error) {
	recip, err := computeRecip(batch, g.Dtype())
	if err != nil {
		return fmt.Errorf("%s: %w", "In divBatch()", err)
	}

	_, err = tensor.Mul(g, recip, tensor.UseUnsafe())
	if err != nil {
		return fmt.Errorf("%s: %w", "Cannot multiply by reciprocal of batch count", err)
	}
	return nil
}

func clipGrad(g tensor.Tensor, clip, negClip any) (err error) {
	if _, err = tensor.Clamp(g, negClip, clip, tensor.UseUnsafe()); err != nil {
		return fmt.Errorf("%s: %w", clampFail, err)
	}
	return nil
}
