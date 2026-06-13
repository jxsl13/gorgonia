//go:build arm64

package asmcheck

import (
	"math/rand"
	"testing"

	"gorgonia.org/vecf32"
	"gorgonia.org/vecf64"
)

const benchN = 8192

func BenchmarkVecf32Add_NEON(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	x, y := randSlice(rng, benchN), randSlice(rng, benchN)
	b.ResetTimer()
	for b.Loop() {
		vecf32.Add(x, y)
	}
}

func BenchmarkVecf32Add_Scalar(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	x, y := randSlice(rng, benchN), randSlice(rng, benchN)
	b.ResetTimer()
	for b.Loop() {
		for j := range x {
			x[j] += y[j]
		}
	}
}

func BenchmarkVecf64Mul_NEON(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	x, y := randSlice64(rng, benchN), randSlice64(rng, benchN)
	b.ResetTimer()
	for b.Loop() {
		vecf64.Mul(x, y)
	}
}

func BenchmarkVecf64Mul_Scalar(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	x, y := randSlice64(rng, benchN), randSlice64(rng, benchN)
	b.ResetTimer()
	for b.Loop() {
		for j := range x {
			x[j] *= y[j]
		}
	}
}
