# SPEC — gorgonia modernize

## §G goal

Move lib to module `github.com/jxsl13/gorgonia` (was `gorgonia.org/gorgonia`).
Bump go directive to latest stable + toolchain. Bump all deps. Add dependabot
to keep deps fresh. No API/behavior change — path + versions only. [Phase 1, done.]

Phase 2 — Apple Silicon perf backends:
- (a) ARM64 NEON assembly for hot math/vector ops on `darwin/arm64`. Pure Go
  plan9 asm, NO cgo, arch-gated, with scalar pure-Go fallback. Mirror existing
  `mathutils_amd64.{s,go}`.
- (b) Metal GPU backend = new `metal/` subpkg implementing `tensor.Engine`,
  `metal` build tag, cgo + Obj-C + MPSGraph (Metal Performance Shaders).
  Mirror existing `cuda/` backend + `*_cuda.go` VM wiring.
- (c) ANE/NPU [Phase 3, committed]. Not op-level programmable; reachable only
  via CoreML. Build a model-export pipeline: gorgonia graph -> CoreML MIL ->
  `.mlpackage` -> CoreML runtime inference (ANE+GPU+CPU). Uses `gomlx/go-coreml`.
  Separate from the VM/tensor.Engine — a model-level inference path.

## §C constraints

- C1: external `gorgonia.org/{cu,tensor,vecf32,vecf64,dawson,golgi}` = upstream,
  not ours. Keep verbatim. ONLY self-module renamed.
- C2: self-module = `gorgonia.org/gorgonia` + subpkgs (`blase`, `cuda`,
  `examples/mnist`, `internal/encoding`, `ops/nn`, `x/vm`). 82 `.go` files.
- C3: `cuda`/`blase` build-tag gated (need CUDA/BLAS) — rename applies,
  compile of those tags may not run in CI. Default-tag build must stay green.
- C4: `.github/dependabot.yml` exists but empty + untracked. Fill it.
- C5: go directive = `1.26`. toolchain = `go1.26.4` (installed latest stable).
- C6: `go4.org/unsafe/assume-no-moving-gc` (indirect via `gorgonia.org/tensor`)
  kept as updated dep — NOT vendored, NOT generated, NO custom CI. Current
  version (T3 bump, `//go:build go1.21`) queries `runtime.heapObjectsCanMove`
  via `//go:linkname` → works on go1.26+ with no panic, no env flag. Old
  pre-1.21 version used generated per-version build-tag files + panicked on
  go1.26; T3 bump resolved it. dependabot (T4) keeps it fresh going forward.

### Phase 2 constraints

- C7: ARM64 asm = Go plan9 assembler `.s` files, arch-gated (`*_arm64.s` +
  `*_arm64.go`), NO cgo. go1.26 `simd/archsimd` covers AMD64 only — ARM64 NEON
  hand-written. Pure-Go scalar fallback MUST exist for other GOARCH/GOOS.
- C8: Metal backend = new `metal/` subpkg, `//go:build metal && darwin && arm64`,
  cgo + Obj-C + MPSGraph. Needs macOS 12+, Apple Silicon, Xcode CLT. Mirrors
  `cuda/` Engine pattern. Default build (no `metal` tag) byte-unaffected.
- C9: ANE/NPU NOT directly programmable. Apple exposes it only via CoreML graph
  partitioning (runtime picks ANE vs GPU vs CPU). Per-op VM dispatch cannot
  target ANE. COMMITTED path (Phase 3): translate a gorgonia `*ExprGraph` ->
  CoreML MIL program -> compile `.mlpackage` -> infer via `gomlx/go-coreml`
  (`ComputeAll` = ANE+GPU+CPU). Model-level, not a tensor.Engine. go-coreml is
  alpha — pin version, isolate behind our own interface. Needs macOS 12+, Xcode. RESOLVED (B6): coremlcompiler (Xcode-only) NOT required — link CoreML.framework directly via cgo and compile at runtime with [MLModel compileModelAtURL:error:] (present with CLT). Compute units via MLModelConfiguration.computeUnits (MLComputeUnitsAll = ANE+GPU+CPU). Drop hard dep on gomlx/go-coreml's coremlcompiler step; may still reuse its MIL protobuf types.
- C10: low-level vector asm (axpy etc.) lives in EXTERNAL `gorgonia.org/vecf32`
  /`vecf64`. To add ARM64 NEON there, VENDOR them per C12 (internal copy + our
  asm on top via `replace`) instead of an upstream PR. See T20. In-repo asm
  (`mathutils`, `nn`) stays in-repo.
- C12: VENDORING POLICY. Any dependency that needs OUR modification is vendored
  as an internal package — a verbatim copy of the upstream LATEST release with
  our changes layered on top. Unmodified deps stay normal external requires (do
  NOT vendor; e.g. go4.org per C6). Layout `internal/vendor/<name>/`. Wire via
  go.mod `replace <upstream> => ./internal/vendor/<name>` so transitive
  importers also resolve to our copy. Preserve upstream LICENSE verbatim; record
  provenance (module path + exact version/commit) so our diff stays re-syncable.
- C13: CoreML path = our own `coreml/` cgo pkg linking `-framework CoreML -framework Foundation`, build tag `coreml && darwin && arm64`. Build a `.mlmodel`/`.mlpackage` on disk, compile at RUNTIME via `[MLModel compileModelAtURL:error:]` (no coremlcompiler/Xcode), load, predict. Compute units selectable (All/CPUAndGPU/CPUOnly). Verified linkable with CLT 26.5 (B6).
- C14: CUDA + BLAS code is build-tag-gated so default tooling on a machine WITHOUT those toolchains never compiles it. CUDA (`//go:build cuda`): whole `cuda/` package, `cmd/cudagen`, `examples/convnet_cuda` (root `*_cuda.go` + `ops/nn/*_cuda.go` already tagged). BLAS (`//go:build blas`): `blase/` package, `examples/stacked_autoencoder`. After this, `go build/vet/generate/staticcheck/govulncheck ./...` need NO `/cuda$` `/blase$` grep-excludes. CUDA/BLAS only build with `-tags cuda`/`-tags blas` on hosts with the toolchain (cannot verify the positive here, C3).
- C15: CI runs ONLY on GitHub-hosted runners (no self-hosted) using the `*-latest` labels (`ubuntu-latest`, `macos-latest`); actions pinned to latest major (`actions/checkout@v5`, `actions/setup-go@v5`, `codecov/codecov-action@v5`), kept fresh by dependabot (T4). NO in-repo workflow generator — static hand-written workflows only. Every workflow + config YAML uses the `.yaml` extension (not `.yml`).
- C11: research refs — Metal: mikecvet/go-mm, tsawler/go-metal (MPSGraph).
  ARM64 NEON: jairad26/go-simd, axiomhq/simd-go, pehringer/simd. ANE: 
  gomlx/go-coreml, fredyshox/ANECompat. See §R.

## §I surfaces

- I.mod: `go.mod` module line + `go`/`toolchain` directives + require block.
- I.imports: all `import "gorgonia.org/gorgonia[/...]"` across repo.
- I.dependabot: `.github/dependabot.yml` — pkg-ecosystems gomod + github-actions.
- I.ci: `.github/workflows/*` — go version matrix may pin old go.

### Phase 2 surfaces

- I.asm: in-repo arch-gated files. Existing `mathutils_amd64.{s,go}`. Add
  `mathutils_arm64.{s,go}` (NEON) + `nn_arm64.{s,go}`. Each asm fn = exported
  Go wrapper + pure-Go scalar fallback (`*_noasm.go` / generic build tag).
- I.metal: new `metal/` subpkg implementing `tensor.Engine` (mirror `cuda/`:
  engine.go, arith.go, linalg.go, arena.go...). Build tag `metal`. Obj-C/MSL
  sources + cgo preamble (link `-framework Metal -framework MetalPerformanceShaders`).
- I.metal-vm: VM wiring `vm_tape_metal.go` / `op_math_metal.go` mirror
  `*_cuda.go`, tag-gated. Example `examples/*_metal`.
- I.coreml: new `coreml/` subpkg (Phase 3) — public API: `Export(g *ExprGraph,
  outputs ...*Node) (*Model, error)`, `Model.Predict(inputs) (outputs, error)`,
  compute-unit selector (ANE+GPU+CPU / CPU+GPU / CPU). Build tag
  `coreml && darwin && arm64`, cgo. Wraps `gomlx/go-coreml`. `.mlpackage` on disk.
- I.vendor: `internal/vendor/<name>/` per vendored dep — verbatim upstream
  source + `UPSTREAM.md` (module path, version/commit, sync date) + LICENSE copy
  + our patch files. go.mod gains a `replace` per vendored dep.
- I.ci-darwin: CI runner darwin/arm64 = GitHub-hosted `macos-14` (Apple Silicon).
  Builds default + `metal` tag (+ `coreml` tag Phase 3), runs asm + metal parity
  tests. Note: hosted GPU/ANE access may be limited — parity tests fall back to
  CPU compute-unit when device unavailable; build/compile always gated.
- I.makefile: root `Makefile` with targets devs run locally — `fmt`, `tidy`, `vet`, `lint` (staticcheck), `vuln` (govulncheck), `test`, `pre-check` (mirror the CI gate), `check` (all). Lets us SEE + FIX issues.

## §V invariants

- V1: only self-module renamed. `gorgonia.org/{cu,tensor,vecf32,vecf64,dawson,golgi}`
  imports byte-identical pre/post.
- V2: post-rename ZERO `gorgonia.org/gorgonia` import strings in OUR `*.go`,
  ZERO in `go.mod` `module` line + direct-require block. EXCEPTION: an
  `// indirect` `gorgonia.org/gorgonia` require is allowed — upstream
  `gorgonia.org/cu/dnn/interop` + `gorgonia.org/golgi` legitimately depend on
  the original (different) module. See B1.
- V3: `go build ./...` (default tags) green. `go vet ./...` green.
- V4: existing test suite green pre vs post — no new failures introduced by rename.
- V5: `dependabot.yml` valid: `version: 2`, watches `gomod` + `github-actions`,
  scheduled (weekly).
- V6: `go.mod` `go 1.26`; `toolchain go1.26.4`.
- V7: behavior unchanged. Edits limited to: path strings, dep versions, go
  version, + minimal call-site adapts forced by major dep API changes (T3).
  No feature/logic redesign. OUR public API identifiers unchanged (only import
  path differs). Each dep-forced edit noted in §B.
- V8: CI workflows go-version updated to match go.mod (no stale old go).
- V9: `go vet ./...` (non-cgo pkgs) reports zero `non-constant format string`.
  Newer toolchain promotes this to build-fail — keep format args constant.
- V10: tests/examples requiring reproducible RNG use a local
  `rand.New(rand.NewSource(seed))`, never global `rand.*` — a dep goroutine
  may consume the global source, and go1.20+ auto-seeds it.

### Phase 2 invariants

- V11: every ARM64 asm fn has a pure-Go scalar fallback chosen by build tags;
  non-arm64 / non-darwin builds compile + produce identical results. Parity
  test asserts asm == scalar (bit-exact for int, documented float tol).
- V12: asm = Go plan9 syntax, NO cgo. `go build`+`go vet` green on darwin/arm64.
- V13: `metal` backend behind `//go:build metal && darwin && arm64`. Default
  build (no tag) byte-identical to pre-Phase2. metal tag compiles only on
  darwin/arm64.
- V14: metal Engine satisfies `tensor.Engine`; matmul/conv/elementwise results
  match CPU engine within documented float tolerance (parity set).
- V15: GPU buffers alloc/free paired; pool released on Engine close — no leak
  (verified by repeated alloc loop + instruments/leak check).
- V16: CoreML export path behind `//go:build coreml && darwin && arm64`; default
  build (no tag) byte-unaffected. `coreml/` exposes our own interface — no
  `gomlx/go-coreml` type leaks into public API (alpha dep isolated).
- V17: CI darwin/arm64 job = GH `macos-14`. Builds default + `metal` (+ `coreml`)
  tags, runs asm parity + metal/coreml parity tests, green. Device-bound tests
  skip cleanly (not fail) when hosted runner lacks GPU/ANE.
- V18: CoreML `Model.Predict` output matches the equivalent gorgonia CPU graph
  within documented float tol on a parity model set. Compute-unit = best-effort
  ANE; CoreML may partition to GPU/CPU — correctness asserted, NOT that it ran
  on the NPU. ANECompat used to report expected ANE residency (advisory).
- V19: ALWAYS quote sources. Every external/research-derived claim (API
  behavior, version support, perf number, platform requirement) carries a
  source URL in §R. New refs append to §R, never inline-only. Reports/PRs that
  cite research link the source.
- V20: every `internal/vendor/<name>/` records provenance (upstream module path
  + exact version/commit + sync date) in `UPSTREAM.md`, and keeps the original
  LICENSE verbatim. Our mods isolatable from the upstream baseline (separate
  files or marked patch) so re-sync stays mechanical.
- V21: vendored copy starts from upstream LATEST release = copy + our diff on
  top, NOT a rewrite. baseline diffable against pristine upstream.
- V22: ONLY deps requiring our modification are vendored. Unmodified deps stay
  normal external requires — no needless vendoring (go4.org stays external, C6).
- V23: vendored-via-`replace` resolves the original import path to the local
  copy; transitive importers (e.g. `tensor` -> `vecf32`) pick up our copy.
  `go mod verify` + build green; pristine build (no vendored mods) parity-tested.
- V24: NO asm added without a T7 bench proving the op is hot AND SIMD-able.
  Cold/scalar ops (e.g. `divmod`) stay generic Go — asm symmetry alone is not a
  reason. Measure-first (see B4).
- V25: commit after each completed + consistent step. When a §T verifies green,
  commit it (`T<n>: <goal>` + §V cites) BEFORE the next task — one task per
  commit, never batch multiple tasks. A split task (e.g. T11->T21) commits at
  the consistent boundary it actually reached.
- V26: a host-accessible GPU tensor.Engine (embeds tensor.StdEng) integrates end-to-end via gorgonia NewTapeMachine(g, WithEngine(e)) — NO cuda-style device-transfer machinery. TapeMachine sets every value's engine to m.Engine (vm_tape.go), so pass the engine to the MACHINE, not only to let-bound values. Non-overridden ops fall back to StdEng on CPU. B7.
- V27: CI pre-check gate enforces, with `git diff --exit-code` after each mutating cmd: `gofmt -l` (exclude internal/vendor — V20 verbatim), `go mod tidy`, `go generate` (exclude `cuda` pkg — only generator is CUDA cudagen needing the toolchain), and `govulncheck` (latest, scoped to non-cgo-lib pkgs). Repo MUST stay gofmt-clean + tidy-clean + vuln-free.
- V28: staticcheck is available via `make lint` (local, exclude internal/vendor) and runs in the CI pre-check as ADVISORY (continue-on-error) — the gorgonia lib carries ~267 pre-existing issues, so it surfaces but does NOT fail CI. Our new packages (coreml/metal/ asmcheck) stay staticcheck-clean. Tighten to blocking once legacy is cleaned. ?
- V29: on a host WITHOUT CUDA/BLAS, `go build ./...` (no tags, no grep-excludes) succeeds — every `gorgonia.org/cu` / CBLAS-importing file is gated by `cuda` / `blas`. CI/Makefile drop the `/cuda$` `/blase$` excludes. `-tags cuda` / `-tags blas` compile the gated code (needs the toolchain).
- V30: `.github/workflows/` holds ONLY static `.yaml` workflows on GitHub-hosted runners — no `runs-on: self-hosted`, no `*.go` generator, no `*.yml`. Final set: `pre-check.yaml`, `linux.yaml`, `darwin-arm64.yaml`, `coverage.yaml`. `.github/dependabot.yml` -> `.yaml`. `grep -r self-hosted .github` empty.

## §T tasks

```
id|status|task|cites
T1|x|rename module: go.mod module line + 82 files import path gorgonia.org/gorgonia -> github.com/jxsl13/gorgonia, leave other gorgonia.org/* untouched|V1,V2,C1,I.imports,I.mod
T2|x|set go.mod go 1.26 + toolchain go1.26.4|V6,C5,I.mod
T3|x|bump deps incl MAJOR upgrades (gota,arrow,protobuf,etc): go get -u then chase major versions, go mod tidy, fix API breakage|V3,V7,I.mod
T4|x|write .github/dependabot.yml: v2, gomod + github-actions, weekly|V5,C4,I.dependabot
T5|x|update CI workflows go-version to match go.mod|V8,I.ci
T6|x|verify: go build ./... + go vet + go test ./... green; confirm zero stale refs|V2,V3,V4
T7|x|bench scalar baseline for hot math/nn ops (mathutils, nn) on darwin/arm64 — profile before any asm|V12,I.asm,C7
T8|x|REDIRECTED (B4): add ARM64 asm ONLY to in-repo ops T7 proves hot+SIMD-able; divmod stays generic Go. If none in-repo, T8 is no-op and win lives in T20/T9|V11,V12,V24,I.asm,C7,C10
T9|x|extend NEON to nn hot ops (activations, conv inner) per T7 profile; parity + bench|V11,V12,I.asm,C7
T10|x|scaffold metal/ subpkg mirroring cuda/: Engine skeleton, build tag metal&&darwin&&arm64, cgo+ObjC+MPSGraph preamble, buffer alloc/free + pool|V13,V15,C8,I.metal
T11|x|impl metal ops: elementwise (MSL kernels) + matmul (MPSMatrixMultiplication); parity vs CPU. Conv split to T21|V14,V15,I.metal
T12|x|example examples/metal (GPU elementwise + matmul, runs on M2 Pro). VM auto-dispatch wiring split to T22|V13,I.metal-vm
T22|x|Metal tensor.Engine (embed StdEng + GPU MatMul via MatMuler); tensors WithEngine(metal.Engine) auto-dispatch MatMul to GPU; parity vs CPU|V14,I.metal-vm
T23|x|full TapeMachine device-transfer wiring: *_metal.go mirror device_cuda.go/op_math_cuda.go/vm_tape_cuda.go so a gorgonia graph runs end-to-end on GPU (large)|V13,V14,I.metal-vm
T24|x|CI pre-check job: gofmt + go mod tidy + go generate + govulncheck, fail on any git diff (V27)|V27,I.ci
T25|x|add staticcheck: root Makefile (fmt/tidy/vet/lint/vuln/test/pre-check/check) + advisory staticcheck step in pre-check.yml (excl vendored); keep our pkgs clean|V28,I.makefile,I.ci
T26|x|//go:build cuda on cuda/ package + cmd/cudagen + examples/convnet_cuda; drop /cuda$ grep-excludes; verify go build ./... clean without excludes on non-CUDA host|V29,C14,I.ci
T27|x|//go:build blas on blase/ + examples/stacked_autoencoder; drop /blase$ grep-excludes; verify go build ./... clean without excludes on non-BLAS host|V29,C14,I.ci
T28|x|delete workflow generator (.github/workflows/main.go + job-template.go) + runner-self-hosted.yml + runner-github-{macos,ubuntu}-amd64.yml|V30,C15,I.ci
T29|x|add static linux.yaml (ubuntu-latest: cross-build arm/amd64/darwin + go test -race + avx/sse tag builds); modernize coverage->coverage.yaml (ubuntu-latest, checkout@v5/setup-go@v5/codecov@v5); darwin on macos-latest; all actions @latest|V30,C15,I.ci
T30|x|rename all .yml -> .yaml (pre-check, darwin-arm64, coverage, .github/dependabot); verify grep -r self-hosted .github empty|V30,C15,I.ci
T13|x|CI darwin/arm64 runner (GH macos-14): build default + metal tag, run asm parity + metal parity tests; device-bound tests skip when no GPU|V17,I.ci-darwin
T14|x|Phase3 spike: gomlx/go-coreml hello-world — load/compile .mlpackage, infer, select compute units; pin alpha version|C9,I.coreml
T15|x|Phase3: coreml/ subpkg + public iface (Export/Model/Predict/compute-unit), build tag coreml&&darwin&&arm64, isolate go-coreml types|V16,C9,I.coreml
T16|x|Phase3: graph->CoreML MIL translator for supported op subset (matmul, conv, activations, pooling, add/mul); unsupported op -> clear error|V18,I.coreml
T17|x|Phase3: parity tests Model.Predict vs CPU graph within tol; ANECompat advisory report; example examples/*_coreml|V18,I.coreml
T18|x|Phase3: extend CI macos-14 to build+test coreml tag (device tests skip when no ANE)|V17,I.ci-darwin
T19|x|establish vendoring convention: internal/vendor/<name>/ layout, UPSTREAM.md provenance template, replace-directive pattern, LICENSE rule|V20,V21,V22,C12,I.vendor
T20|x|vendor gorgonia.org/vecf32 + vecf64 (latest) into internal/vendor; add ARM64 NEON asm + scalar fallback; replace directives; parity asm==scalar; transitive tensor picks up copy|V11,V20,V21,V23,C10,C12,I.vendor,I.asm
T21|x|metal GPU conv2d via MPSGraph convolution2D; parity vs CPU nn conv (split from T11)|V14,V15,I.metal
```

## §B bugs

```
id|date|cause|fix
B1|2026-06-13|after rename go.mod regained gorgonia.org/gorgonia // indirect — upstream cu/dnn/interop + golgi depend on original module, not removable|V2 (allow indirect-only exception)
B2|2026-06-13|go1.26 vet promotes non-constant format string to build-fail; 8 WriteHash sites fmt.Fprintf(h, op.String()) blocked go test|fixed -> fmt.Fprintf(h, "%s", op.String()); V9
B3|2026-06-13|Example_linearRegression flaky after dep bump: used global math/rand.Float*; a bumped dep spawns goroutine consuming global rand -> dataset nondeterministic -> Output mismatch (old deps masked it)|xy()/random() use local rand.New(rand.NewSource(seed)); seed1 reproduces prior sequence, Output unchanged; V10
B4|2026-06-13|T8 assumed in-repo mathutils had SIMD-able hot ops; mathutils = only divmod (scalar int div, cold: shape-infer/bitmap/ctc). Not a NEON candidate; arm64 Go already emits UDIV+MSUB|redirect T8 to profile-driven targets; divmod stays generic Go; V24
B5|2026-06-13|T14-T18 (CoreML/ANE) blocked on dev machine: gomlx/go-coreml needs coremlcompiler = FULL Xcode; only Command Line Tools present (xcrun cannot find coremlcompiler). CoreML code cannot be built/verified here|defer T14-T18 to a full-Xcode env; tasks stay . (blocked), not faked; C9 amended
B6|2026-06-13|B5 reassessed: coremlcompiler only needed for OFFLINE .mlpackage compile. CoreML.framework runtime API [MLModel compileModelAtURL:error:] works with CLT-only (probe confirmed) -> Xcode NOT required|unblock T14-T18 via direct cgo CoreML.framework + runtime compile; C9 amended; C13 added
B7|2026-06-13|T23 assumed a ~4000-line mirror of cuda device-transfer machinery. WRONG: cuda needs that only because CUDA memory is NOT host-accessible. Metal engine embeds StdEng (host-accessible) -> plugs into gorgonia NewTapeMachine(g, WithEngine(e)). First test got 0 GPU dispatches: machine overrides value engines with m.Engine (default StandardEngine)|pass metal engine via WithEngine; no new VM files; V26
```

## §R refs

Phase 2 research (2026-06):
- Metal GPU from Go: [mikecvet/go-mm](https://github.com/mikecvet/go-mm)
  (cgo+ObjC+MPS+MSL, benchmarks vs gonum/OpenBLAS),
  [tsawler/go-metal](https://github.com/tsawler/go-metal) (MPSGraph deep-learning
  lib, macOS 12+/Apple Silicon).
- ARM64 NEON asm (no cgo): [jairad26/go-simd](https://github.com/jairad26/go-simd),
  [axiomhq/simd-go](https://github.com/axiomhq/simd-go) (NEON+SVE, threshold
  dispatch), [pehringer/simd](https://github.com/pehringer/simd).
  go1.26 `simd/archsimd` = AMD64 only; ARM64 via hand asm
  ([ajroetker/go-highway](https://github.com/ajroetker/go-highway)).
- ANE/NPU: [gomlx/go-coreml](https://github.com/gomlx/go-coreml) (alpha, CoreML
  bindings, ComputeAll = ANE+GPU+CPU), [fredyshox/ANECompat](https://github.com/fredyshox/ANECompat)
  (ANE op-compat checker). ANE reachable only through CoreML graph partition.
