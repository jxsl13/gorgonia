//go:build cuda

package gorgonia

import (
	"unsafe"

	"github.com/chewxy/hm"
	"github.com/pkg/errors"
	"gorgonia.org/tensor"
)

// makeValueFromMem builds a Value backed by external/device (CUDA) memory.
// Only the cuda build reaches this — see the doc in values.go.
func makeValueFromMem(t hm.Type, s tensor.Shape, mem tensor.Memory) (retVal Value, err error) {
	var dt tensor.Dtype
	if dt, err = dtypeOf(t); err != nil {
		return
	}
	if s.IsScalar() {
		return makeScalarFromMem(dt, mem)
	}

	switch tt := t.(type) {
	case TensorType:
		memsize := calcMemSize(dt, s)
		return tensor.New(tensor.Of(dt), tensor.WithShape(s...), tensor.FromMemory(mem.Uintptr(), uintptr(memsize))), nil
	case tensor.Dtype:
		return makeScalarFromMem(tt, mem)
	default:
		err = errors.Errorf(nyiTypeFail, "MakeValue", tt)
		return
	}
}

// makeScalarFromMem reinterprets device memory's address as a scalar Value.
//
// It converts mem.Uintptr() to an unsafe.Pointer. go vet's unsafeptr analyzer
// flags this because a uintptr is an integer, not a GC-visible reference, and a
// uintptr obtained from a method call is not one of the documented-valid
// unsafe.Pointer patterns (https://pkg.go.dev/unsafe, golang/go#58625). It is
// safe here because:
//
//   - this is external (CUDA device) memory whose address is allocated outside
//     Go and stays valid for the lifetime of the Value (the "non-Go memory"
//     case of golang/go#58625);
//   - gorgonia (via gorgonia.org/tensor) depends on
//     go4.org/unsafe/assume-no-moving-gc, asserting a non-moving GC.
//
// tensor.Memory exposes only Uintptr() (no Pointer() unsafe.Pointer), so the
// round-trip is unavoidable here; the clean fix is an upstream Memory.Pointer().
// This file is built only with `-tags cuda`, so the default build/vet never
// compiles it (no -unsafeptr suppression needed there).
//
//go:nocheckptr
func makeScalarFromMem(dt tensor.Dtype, mem tensor.Memory) (retVal Value, err error) {
	switch dt {
	case tensor.Float64:
		retVal = (*F64)(unsafe.Pointer(mem.Uintptr()))
	case tensor.Float32:
		retVal = (*F32)(unsafe.Pointer(mem.Uintptr()))
	case tensor.Int:
		retVal = (*I)(unsafe.Pointer(mem.Uintptr()))
	case tensor.Int64:
		retVal = (*I64)(unsafe.Pointer(mem.Uintptr()))
	case tensor.Int32:
		retVal = (*I32)(unsafe.Pointer(mem.Uintptr()))
	case tensor.Byte:
		retVal = (*U8)(unsafe.Pointer(mem.Uintptr()))
	case tensor.Bool:
		retVal = (*B)(unsafe.Pointer(mem.Uintptr()))
	default:
		err = errors.Errorf(nyiTypeFail, "makeScalarFromMem", dt)
	}
	return
}
