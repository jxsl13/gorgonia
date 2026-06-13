//go:build cuda
// +build cuda

package cuda

//go:generate cudagen

const (
	elemBinOpMod   = "elembinop"
	elemUnaryOpMod = "elemunaryop"
)
