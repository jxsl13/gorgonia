//go:build metal && darwin && arm64
// +build metal,darwin,arm64

package metal

import (
	"math"
	"math/rand"
	"testing"
)

func cpuMatMul(a, b []float32, m, n, k int) []float32 {
	c := make([]float32, m*n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			var s float32
			for p := 0; p < k; p++ {
				s += a[i*k+p] * b[p*n+j]
			}
			c[i*n+j] = s
		}
	}
	return c
}

func TestMetalMatMulParity(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Skipf("metal unavailable: %v", err)
	}
	defer d.Close()

	rng := rand.New(rand.NewSource(1))
	shapes := [][3]int{{1, 1, 1}, {2, 3, 4}, {16, 16, 16}, {32, 64, 48}, {128, 128, 128}}
	for _, s := range shapes {
		m, n, k := s[0], s[1], s[2]
		a := make([]float32, m*k)
		b := make([]float32, k*n)
		for i := range a {
			a[i] = rng.Float32()*2 - 1
		}
		for i := range b {
			b[i] = rng.Float32()*2 - 1
		}
		got, err := d.MatMul(a, b, m, n, k)
		if err != nil {
			t.Fatalf("MatMul %v: %v", s, err)
		}
		want := cpuMatMul(a, b, m, n, k)
		// GPU summation order differs -> relative tolerance, not bit-exact (V14).
		const tol = 1e-4
		for i := range want {
			diff := math.Abs(float64(got[i] - want[i]))
			denom := math.Max(1, math.Abs(float64(want[i])))
			if diff/denom > tol {
				t.Fatalf("MatMul %v idx=%d: gpu=%v cpu=%v (rel %g > %g)", s, i, got[i], want[i], diff/denom, tol)
			}
		}
	}
}
