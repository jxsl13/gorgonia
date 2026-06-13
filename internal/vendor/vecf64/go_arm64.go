// MOD(jxsl13): ARM64 path. Add/Sub/Mul use NEON (add_arm64.s); Div/Sqrt/InvSqrt
// stay pure-Go (Div forces divide-by-zero -> +Inf; Sqrt/InvSqrt use math).
// T20, SPEC §V11.

//go:build arm64
// +build arm64

package vecf64

import "math"

//go:noescape
func addAsm(a, b []float64)

//go:noescape
func subAsm(a, b []float64)

//go:noescape
func mulAsm(a, b []float64)

// Add performs a̅ + b̅. a̅ will be clobbered
func Add(a, b []float64) {
	if len(a) != len(b) {
		panic("vectors must be the same length")
	}
	addAsm(a, b)
}

// Sub performs a̅ - b̅. a̅ will be clobbered
func Sub(a, b []float64) {
	if len(a) != len(b) {
		panic("vectors must be the same length")
	}
	subAsm(a, b)
}

// Mul performs a̅ × b̅. a̅ will be clobbered
func Mul(a, b []float64) {
	if len(a) != len(b) {
		panic("vectors must be the same length")
	}
	mulAsm(a, b)
}

// Div performs a̅ ÷ b̅. a̅ will be clobbered
func Div(a, b []float64) {
	b = b[:len(a)]
	for i, v := range a {
		if b[i] == 0 {
			a[i] = math.Inf(0)
			continue
		}
		a[i] = v / b[i]
	}
}

// Sqrt performs √a̅ elementwise. a̅ will be clobbered
func Sqrt(a []float64) {
	for i, v := range a {
		a[i] = math.Sqrt(v)
	}
}

// InvSqrt performs 1/√a̅ elementwise. a̅ will be clobbered
func InvSqrt(a []float64) {
	for i, v := range a {
		a[i] = float64(1) / math.Sqrt(v)
	}
}
