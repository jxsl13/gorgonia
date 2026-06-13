# ANE residency (advisory)

The CoreML runtime partitions a model across the **Apple Neural Engine (ANE),
GPU, and CPU** and decides per-op where to run. `coreml.Export` requests
`ComputeAll` by default, but ANE residency is **not guaranteed** — CoreML may
fall back to GPU/CPU for ops the ANE doesn't support (SPEC §V18).

Our parity tests assert **correctness**, not that a given op ran on the ANE.

## Checking expected ANE residency

To inspect which ops are expected to land on the ANE, run the produced
`.mlmodelc` / `.mlpackage` through the external tool **ANECompat**:

- <https://github.com/fredyshox/ANECompat>

It reports whether a model runs end-to-end on the ANE or only in segments, and
flags ops that force a fallback. Use it when designing a model architecture you
want ANE-resident.

## Op subset → ANE-friendliness (current translator)

| op (T16) | typically ANE-eligible |
|---|---|
| matmul | yes (as a 1x1 conv / inner-product on ANE) |
| add / sub / mul / div (elementwise) | yes |

Unsupported gorgonia ops error at `Export` time rather than silently miscompile.
