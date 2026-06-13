# Phase 2 — ARM64 perf baseline (T7)

Measure-first gate before any assembly (SPEC §V24, §C7). All numbers
`darwin/arm64`, Apple Silicon, go1.26.4, `-benchtime=1x`.

## Baseline benchmarks

| bench | ns/op | B/op | allocs/op |
|---|---|---|---|
| Reshape_Dense/simple | 591,667 | 38,064 | 309 |
| Reshape_Dense/simple_big_tensor | 248,583 | 358,232 | 274 |
| GroupNorm | 1,049,000 | 120,864 | 1,376 |
| SoftmaxLargeOldAxis0 | 35,499,012,542 | 13,103,973,576 | 125,863,289 |

## CPU profile (SoftmaxLargeOldAxis0)

```
flat   flat%   cum    function
10.06s 10.06%  10.69s github.com/jxsl13/gorgonia.Uniform64           (in-repo)
 7.05s  7.05%   9.77s gorgonia.org/tensor.StdEng.softMaxInnerDimF64.func1 (external)
 0.09s  0.09%   2.36s gorgonia.org/tensor.StdEng.softMaxInnerDimF64       (external)
```

## Findings

1. **In-repo SIMD-able hot loops are scarce.** The heavy numeric kernels
   (softmax, matmul, conv) live in **external `gorgonia.org/tensor`**, not in
   this module. Optimizing them = vendor `tensor`/`vecf32`/`vecf64` (SPEC §C12,
   T20) or the Metal backend (T10) — NOT in-repo `.s` files.
2. **`Uniform64` dominates micro-benchmarks** (10s flat) but is a sequential
   PRNG fill (`weights.go:291`, leesper/go_rng) — weight-init setup, called once
   per graph build, and a poor NEON target (each draw depends on prior PRNG
   state). Not an asm candidate.
3. **`mathutils` = `divmod` only** — scalar int division, cold (shape-infer /
   bitmap / ctc). Confirms B4: no in-repo NEON target. arm64 Go already emits
   `UDIV`+`MSUB`.

## Decision

- **T8** (in-repo `mathutils_arm64` NEON): no-op. No in-repo op clears the
  hot + SIMD-able bar (B4, V24).
- **ARM64 asm effort → T20**: vendor `vecf32`/`vecf64`, add NEON to the vector
  primitives (axpy, scale, etc.) where the real per-element work happens.
- **T9** (nn NEON): re-evaluate against `op_nn.go` raw-loop ops
  (im2col/maxpool/batchnorm) only if a profile of a real training/inference
  workload (not weight-init) shows them hot.

## T9 — real-workload profile (nn)

`BenchmarkTrainingNonConcurrent` (38ms/op) + `BenchmarkTapeMachineExecution`
(27ms/op), `-benchtime=200x`. Top compute funcs:

```
cum%    function
 9.93%  github.com/jxsl13/gorgonia.StandardEngine.Transpose   (in-repo)
 6.20%  gorgonia.org/tensor.(*FlatIterator).Next              (external)
 3.33%  gorgonia.org/tensor.(*FlatIterator).ndNext            (external)
 2.07%  gorgonia.org/tensor.StdEng.MulScalar                  (external)
```

Findings:
- Hot path = tensor iteration + `Transpose` + std-engine scalar ops. The
  arithmetic kernels are in **external `tensor`**.
- The one in-repo hotspot, `StandardEngine.Transpose`, is a memory-bound
  permutation/shuffle — NOT a NEON arithmetic target.
- `op_nn` raw loops (im2col/maxpool/batchnorm) did not surface in the profile.

Decision: **T9 = no in-repo NEON target** (V24). Confirms B4 for nn. ARM64 perf
surface is external `tensor`/`vecf32`/`vecf64` → vendoring (T20) or Metal (T10).

## T20 — vendored vecf32/vecf64 NEON (done)

Vendored `gorgonia.org/vecf32` + `vecf64` @v0.9.0 into `internal/vendor/`
(SPEC §C12), added ARM64 NEON `Add`/`Sub`/`Mul` (WORD-encoded FADD/FSUB/FMUL,
4×f32 / 2×f64 per iter), `Div`/`Sqrt`/`InvSqrt` stay pure-Go. Bit-exact parity
vs scalar (`internal/asmcheck`, lengths 0–33). Transitive `tensor`→`vecf32/64`
resolves to our copy (`replace`, §V23).

| op | NEON | scalar | speedup |
|---|---|---|---|
| vecf32 Add (8192) | 958 ns | 4281 ns | 4.5× |
| vecf64 Mul (8192) | 1692 ns | 4186 ns | 2.5× |

Full lib suite green incl vendored copies; `go mod verify` clean.

Sources: profiling via `go tool pprof`; vendoring rationale SPEC §C12;
ARM64 NEON encodings runtime-verified.
