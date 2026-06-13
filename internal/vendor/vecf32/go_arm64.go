// MOD(jxsl13): ARM64 path. Add/Sub/Mul use NEON (add_arm64.s); Div/Sqrt/InvSqrt
// stay pure-Go (Div has a divide-by-zero -> +Inf special case that the scalar
// reference forces; Sqrt/InvSqrt use math32). T20, SPEC §V11.

//go:build arm64
// +build arm64

package vecf32

import "github.com/chewxy/math32"

//go:noescape
func addAsm(a, b []float32)

//go:noescape
func subAsm(a, b []float32)

//go:noescape
func mulAsm(a, b []float32)

// Add performs a̅ + b̅. a̅ will be clobbered
func Add(a, b []float32) {
	if len(a) != len(b) {
		panic("vectors must be the same length")
	}
	addAsm(a, b)
}

// Sub performs a̅ - b̅. a̅ will be clobbered
func Sub(a, b []float32) {
	if len(a) != len(b) {
		panic("vectors must be the same length")
	}
	subAsm(a, b)
}

// Mul performs a̅ × b̅. a̅ will be clobbered
func Mul(a, b []float32) {
	if len(a) != len(b) {
		panic("vectors must be the same length")
	}
	mulAsm(a, b)
}

// Div performs a̅ ÷ b̅. a̅ will be clobbered
func Div(a, b []float32) {
	b = b[:len(a)]
	for i, v := range a {
		if b[i] == 0 {
			a[i] = math32.Inf(0)
			continue
		}
		a[i] = v / b[i]
	}
}

// Sqrt performs √a̅ elementwise. a̅ will be clobbered
func Sqrt(a []float32) {
	for i, v := range a {
		a[i] = math32.Sqrt(v)
	}
}

// InvSqrt performs 1/√a̅ elementwise. a̅ will be clobbered
func InvSqrt(a []float32) {
	for i, v := range a {
		a[i] = float32(1) / math32.Sqrt(v)
	}
}
