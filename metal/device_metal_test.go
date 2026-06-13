//go:build metal && darwin && arm64
// +build metal,darwin,arm64

package metal

import (
	"math"
	"math/rand"
	"testing"
)

func TestMetalElementwiseParity(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Skipf("metal unavailable: %v", err)
	}
	defer d.Close()
	t.Logf("GPU: %s", d.Name())

	rng := rand.New(rand.NewSource(1))
	for _, n := range []int{0, 1, 7, 64, 1000, 8193} {
		a := make([]float32, n)
		b := make([]float32, n)
		for i := range a {
			a[i] = rng.Float32()*200 - 100
			b[i] = rng.Float32()*200 - 100
		}
		gotAdd, _ := d.Add(a, b)
		gotSub, _ := d.Sub(a, b)
		gotMul, _ := d.Mul(a, b)
		for i := range a {
			checkF32(t, "Add", n, i, gotAdd[i], a[i]+b[i])
			checkF32(t, "Sub", n, i, gotSub[i], a[i]-b[i])
			checkF32(t, "Mul", n, i, gotMul[i], a[i]*b[i])
		}
	}
}

func checkF32(t *testing.T, op string, n, i int, got, want float32) {
	t.Helper()
	if math.Float32bits(got) != math.Float32bits(want) {
		t.Fatalf("%s n=%d idx=%d: gpu=%v cpu=%v", op, n, i, got, want)
	}
}
