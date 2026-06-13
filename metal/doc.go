// Package metal is a Metal (Apple Silicon GPU) compute backend for gorgonia.
//
// It is compiled only with the `metal` build tag on darwin/arm64 and links the
// Metal + Foundation frameworks via cgo (SPEC §C8). The default build (no tag)
// never compiles this package, so non-Apple-Silicon targets are unaffected.
//
// Build/run its tests with:
//
//	go test -tags metal ./metal/...
//
// SPEC: T10 (scaffold + buffers), T11 (ops), §V13–V15.
package metal
