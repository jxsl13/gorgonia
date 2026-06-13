//go:build native
// +build native

package main

import (
	. "github.com/jxsl13/gorgonia"
	"gonum.org/v1/gonum/blas/gonum"
)

func init() {
	Use(gonum.Implementation{})
}
