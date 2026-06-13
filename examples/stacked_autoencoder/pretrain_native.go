//go:build !blas

package main

import (
	. "github.com/jxsl13/gorgonia"
	"gonum.org/v1/gonum/blas/gonum"
)

func init() {
	Use(gonum.Implementation{})
}
