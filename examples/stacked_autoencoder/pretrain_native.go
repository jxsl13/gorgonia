// +build native

package main

import (
	"gonum.org/v1/gonum/blas/gonum"
	. "github.com/jxsl13/gorgonia"
)

func init() {
	Use(gonum.Implementation{})
}
