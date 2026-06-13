// Parity test for the vendored ARM64 NEON vecf32 ops (SPEC §V11, T20):
// asm Add/Sub/Mul must equal the scalar reference, bit-exact, across lengths
// that exercise both the 4-wide NEON body and the scalar remainder (len%4).

//go:build arm64

package asmcheck

import (
	"math"
	"math/rand"
	"testing"

	"gorgonia.org/vecf32"
)

func refAdd(a, b []float32) {
	for i := range a {
		a[i] += b[i]
	}
}
func refSub(a, b []float32) {
	for i := range a {
		a[i] -= b[i]
	}
}
func refMul(a, b []float32) {
	for i := range a {
		a[i] *= b[i]
	}
}

func randSlice(rng *rand.Rand, n int) []float32 {
	s := make([]float32, n)
	for i := range s {
		s[i] = rng.Float32()*200 - 100
	}
	return s
}

func bitEq(t *testing.T, op string, n int, got, want []float32) {
	t.Helper()
	for i := range got {
		if math.Float32bits(got[i]) != math.Float32bits(want[i]) {
			t.Fatalf("%s n=%d idx=%d: asm=%v scalar=%v (not bit-exact)", op, n, i, got[i], want[i])
		}
	}
}

func TestVecf32NEONParity(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	// lengths 0..33 cover remainder 0,1,2,3 + multiple NEON blocks
	for n := 0; n <= 33; n++ {
		a := randSlice(rng, n)
		b := randSlice(rng, n)

		aAdd, aRef := append([]float32(nil), a...), append([]float32(nil), a...)
		vecf32.Add(aAdd, b)
		refAdd(aRef, b)
		bitEq(t, "Add", n, aAdd, aRef)

		aSub, aRefS := append([]float32(nil), a...), append([]float32(nil), a...)
		vecf32.Sub(aSub, b)
		refSub(aRefS, b)
		bitEq(t, "Sub", n, aSub, aRefS)

		aMul, aRefM := append([]float32(nil), a...), append([]float32(nil), a...)
		vecf32.Mul(aMul, b)
		refMul(aRefM, b)
		bitEq(t, "Mul", n, aMul, aRefM)
	}
}
