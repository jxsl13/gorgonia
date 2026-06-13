// Example: run elementwise + matmul on the Apple Silicon GPU via the metal
// backend. Build/run with the `metal` tag on darwin/arm64:
//
//	go run -tags metal ./examples/metal
//
// SPEC: T12 (example), §C8.

//go:build metal && darwin && arm64
// +build metal,darwin,arm64

package main

import (
	"fmt"
	"log"

	"github.com/jxsl13/gorgonia/metal"
)

func main() {
	d, err := metal.New()
	if err != nil {
		log.Fatalf("metal device: %v", err)
	}
	defer d.Close()
	fmt.Println("GPU:", d.Name())

	// Elementwise add on the GPU.
	a := []float32{1, 2, 3, 4}
	b := []float32{10, 20, 30, 40}
	sum, err := d.Add(a, b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("a + b =", sum)

	// 2x3 · 3x2 matrix multiply on the GPU.
	A := []float32{1, 2, 3, 4, 5, 6}    // 2x3
	B := []float32{7, 8, 9, 10, 11, 12} // 3x2
	C, err := d.MatMul(A, B, 2, 2, 3)   // M=2, N=2, K=3
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("A · B =", C) // [58 64 139 154]
}
