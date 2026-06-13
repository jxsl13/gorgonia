// Package coreml is a CoreML inference path for gorgonia on Apple Silicon.
//
// CoreML can dispatch to the Apple Neural Engine (ANE), GPU, or CPU; the engine
// partitions the model graph and picks per-op (SPEC §C9). This package is a
// model-level inference path — NOT a tensor.Engine — built on a MIL program
// that CoreML compiles and runs.
//
// It compiles models at RUNTIME via the CoreML framework
// ([MLModel compileModelAtURL:error:]), so it needs only the Command Line Tools,
// not a full Xcode / coremlcompiler (SPEC §C13, B6).
//
// Compiled only with the `coreml` build tag on darwin/arm64. Build/run tests:
//
//	go test -tags coreml ./coreml/...
//
// SPEC: T14 (spike), T15–T17 (export pipeline), §V16/V18.
package coreml
