// Parity test for the vendored ARM64 NEON vecf64 ops (SPEC §V11, T20):
// asm Add/Sub/Mul must equal the scalar reference, bit-exact, across lengths
// that exercise both the 2-wide NEON body and the scalar remainder (len%2).

//go:build arm64

package asmcheck

import (
	"math"
	"math/rand"
	"testing"

	"gorgonia.org/vecf64"
)

func refAdd64(a, b []float64) {
	for i := range a {
		a[i] += b[i]
	}
}
func refSub64(a, b []float64) {
	for i := range a {
		a[i] -= b[i]
	}
}
func refMul64(a, b []float64) {
	for i := range a {
		a[i] *= b[i]
	}
}

func randSlice64(rng *rand.Rand, n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = rng.Float64()*200 - 100
	}
	return s
}

func bitEq64(t *testing.T, op string, n int, got, want []float64) {
	t.Helper()
	for i := range got {
		if math.Float64bits(got[i]) != math.Float64bits(want[i]) {
			t.Fatalf("%s n=%d idx=%d: asm=%v scalar=%v (not bit-exact)", op, n, i, got[i], want[i])
		}
	}
}

func TestVecf64NEONParity(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for n := 0; n <= 33; n++ {
		a := randSlice64(rng, n)
		b := randSlice64(rng, n)

		aAdd, aRef := append([]float64(nil), a...), append([]float64(nil), a...)
		vecf64.Add(aAdd, b)
		refAdd64(aRef, b)
		bitEq64(t, "Add", n, aAdd, aRef)

		aSub, aRefS := append([]float64(nil), a...), append([]float64(nil), a...)
		vecf64.Sub(aSub, b)
		refSub64(aRefS, b)
		bitEq64(t, "Sub", n, aSub, aRefS)

		aMul, aRefM := append([]float64(nil), a...), append([]float64(nil), a...)
		vecf64.Mul(aMul, b)
		refMul64(aRefM, b)
		bitEq64(t, "Mul", n, aMul, aRefM)
	}
}
