package gorgonia

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDivmodEdgeCases pins divmod's contract with explicit, hardcoded expected
// results across every edge case. The expectations are written out by hand (Go
// truncates toward zero, so the remainder takes the sign of the dividend); a
// final cross-check asserts they agree with Go's own / and %. This guards any
// future arch-specific implementation (e.g. arm64 SDIV/MSUB).
func TestDivmodEdgeCases(t *testing.T) {
	cases := []struct {
		name         string
		a, b         int
		wantQ, wantR int
	}{
		{"positive", 17, 5, 3, 2},
		{"negative dividend", -17, 5, -3, -2},
		{"negative divisor", 17, -5, -3, 2},
		{"both negative", -17, -5, 3, -2},
		{"zero dividend", 0, 7, 0, 0},
		{"divide by one", 42, 1, 42, 0},
		{"divide by minus one", 42, -1, -42, 0},
		{"dividend < divisor", 3, 10, 0, 3},
		{"neg dividend < divisor", -3, 10, 0, -3},
		{"exact", 100, 10, 10, 0},
		{"large", 1000000, 13, 76923, 1},
		{"maxint / 1", math.MaxInt, 1, math.MaxInt, 0},
		{"minint / 1", math.MinInt, 1, math.MinInt, 0},
		{"minint / -1 (overflow wraps, no panic)", math.MinInt, -1, math.MinInt, 0},
	}

	for _, c := range cases {
		q, r := divmod(c.a, c.b)
		if q != c.wantQ || r != c.wantR {
			t.Errorf("%s: divmod(%d, %d) = (%d, %d); want (%d, %d)", c.name, c.a, c.b, q, r, c.wantQ, c.wantR)
		}
		// cross-check the hardcoded expectation against the language itself
		if gq, gr := c.a/c.b, c.a%c.b; c.wantQ != gq || c.wantR != gr {
			t.Errorf("%s: explicit expectation (%d, %d) disagrees with Go a/b,a%%b = (%d, %d)", c.name, c.wantQ, c.wantR, gq, gr)
		}
	}

	// divide by zero must panic with the standard integer-divide-by-zero panic
	assert.New(t).Panics(func() { divmod(7, 0) })
}

func TestDivmod(t *testing.T) {
	as := []int{0, 1, 2, 3, 4, 5}
	bs := []int{1, 2, 3, 3, 2, 3}
	qs := []int{0, 0, 0, 1, 2, 1}
	rs := []int{0, 1, 2, 0, 0, 2}

	for i, a := range as {
		b := bs[i]
		eq := qs[i]
		er := rs[i]

		q, r := divmod(a, b)
		if q != eq {
			t.Errorf("Expected %d / %d to equal %d. Got %d instead", a, b, eq, q)
		}
		if r != er {
			t.Errorf("Expected %d %% %d to equal %d. Got %d instead", a, b, er, r)
		}
	}

	assert := assert.New(t)
	fail := func() {
		divmod(1, 0)
	}
	assert.Panics(fail)
}
