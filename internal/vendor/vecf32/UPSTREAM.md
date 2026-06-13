# UPSTREAM provenance

- **Upstream module:** `gorgonia.org/vecf32`
- **Upstream repo:** https://github.com/gorgonia/vecf32
- **Vendored version:** `v0.9.0`
- **Vendored commit:** tag `v0.9.0` (latest release as of sync)
- **Sync date:** 2026-06-13
- **License:** MIT (Copyright (c) 2017 Chewxy) — copy preserved in `./LICENSE`
- **go.mod replace:** `replace gorgonia.org/vecf32 => ./internal/vendor/vecf32`

## Our modifications

| file | change | reason | SPEC ref |
|---|---|---|---|
| `add_arm64.s` | added | ARM64 NEON elementwise Add/Sub/Mul/Div | T20, §V11 |
| `asm_arm64.go` | added | arm64 asm fn decls + exported wrappers | T20 |
| `go.go` | MOD: build tag `!avx,!sse` → `!avx,!sse,!arm64` | route arm64 to NEON, not pure-Go | T20 |

## Re-sync notes

Upstream asm is opt-in via `sse`/`avx` build tags; default (incl arm64) uses
pure-Go `go.go`. Our change makes NEON the DEFAULT on arm64 by excluding arm64
from `go.go`'s constraint and adding `asm_arm64.go` + `add_arm64.s`. On re-sync,
re-apply the one-line `go.go` build-tag edit + re-copy the two added files.
