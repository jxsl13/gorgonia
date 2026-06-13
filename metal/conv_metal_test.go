//go:build metal && darwin && arm64

package metal

import (
	"math"
	"math/rand"
	"testing"
)

// cpuConv2D: NHWC src, HWIO weights, zero padding for out-of-bounds.
func cpuConv2D(src, w []float32, n, h, ww, cin, kh, kw, cout, sx, sy, pt, pl, oh, ow int) []float32 {
	out := make([]float32, n*oh*ow*cout)
	for bn := 0; bn < n; bn++ {
		for oy := 0; oy < oh; oy++ {
			for ox := 0; ox < ow; ox++ {
				for oc := 0; oc < cout; oc++ {
					var acc float32
					for ky := 0; ky < kh; ky++ {
						iy := oy*sy - pt + ky
						if iy < 0 || iy >= h {
							continue
						}
						for kx := 0; kx < kw; kx++ {
							ix := ox*sx - pl + kx
							if ix < 0 || ix >= ww {
								continue
							}
							for ci := 0; ci < cin; ci++ {
								s := src[((bn*h+iy)*ww+ix)*cin+ci]
								wt := w[((ky*kw+kx)*cin+ci)*cout+oc]
								acc += s * wt
							}
						}
					}
					out[((bn*oh+oy)*ow+ox)*cout+oc] = acc
				}
			}
		}
	}
	return out
}

func TestMetalConv2DParity(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Skipf("metal unavailable: %v", err)
	}
	defer d.Close()

	rng := rand.New(rand.NewSource(1))
	type tc struct {
		n, h, w, cin, kh, kw, cout, sx, sy, pt, pb, pl, pr int
	}
	cases := []tc{
		{1, 5, 5, 1, 3, 3, 1, 1, 1, 0, 0, 0, 0}, // valid, single channel
		{1, 7, 7, 3, 3, 3, 4, 1, 1, 1, 1, 1, 1}, // same padding, multi-channel
		{2, 8, 6, 2, 3, 3, 5, 2, 2, 0, 0, 0, 0}, // stride 2, batch 2
	}
	for ci, c := range cases {
		src := make([]float32, c.n*c.h*c.w*c.cin)
		wts := make([]float32, c.kh*c.kw*c.cin*c.cout)
		for i := range src {
			src[i] = rng.Float32()*2 - 1
		}
		for i := range wts {
			wts[i] = rng.Float32()*2 - 1
		}
		got, oh, ow, err := d.Conv2D(src, wts, c.n, c.h, c.w, c.cin, c.kh, c.kw, c.cout, c.sx, c.sy, c.pt, c.pb, c.pl, c.pr)
		if err != nil {
			t.Fatalf("case %d: %v", ci, err)
		}
		want := cpuConv2D(src, wts, c.n, c.h, c.w, c.cin, c.kh, c.kw, c.cout, c.sx, c.sy, c.pt, c.pl, oh, ow)
		if len(got) != len(want) {
			t.Fatalf("case %d: len %d != %d", ci, len(got), len(want))
		}
		const tol = 1e-3
		for i := range want {
			diff := math.Abs(float64(got[i] - want[i]))
			denom := math.Max(1, math.Abs(float64(want[i])))
			if diff/denom > tol {
				t.Fatalf("case %d idx=%d: gpu=%v cpu=%v (rel %g > %g)", ci, i, got[i], want[i], diff/denom, tol)
			}
		}
	}
}
