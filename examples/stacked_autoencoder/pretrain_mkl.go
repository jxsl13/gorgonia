//go:build blas
// +build blas

package main

import (
	. "github.com/jxsl13/gorgonia"
	"github.com/jxsl13/gorgonia/blase"
)

func init() {
	Use(blase.Implementation())
}
