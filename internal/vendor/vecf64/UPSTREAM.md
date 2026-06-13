# UPSTREAM provenance

- **Upstream module:** `gorgonia.org/vecf64`
- **Upstream repo:** https://github.com/gorgonia/vecf64
- **Vendored version:** `v0.9.0`
- **Vendored commit:** tag `v0.9.0` (latest release as of sync)
- **Sync date:** 2026-06-13
- **License:** MIT (Copyright (c) 2017 Chewxy) — copy preserved in `./LICENSE`
- **go.mod replace:** `replace gorgonia.org/vecf64 => ./internal/vendor/vecf64`

## Our modifications

| file | change | reason | SPEC ref |
|---|---|---|---|
| `add_arm64.s` | added | ARM64 NEON elementwise Add/Sub/Mul (`.D2`) | T20, §V11 |
| `go_arm64.go` | added | arm64 asm decls + wrappers; Div/Sqrt/InvSqrt pure-Go | T20 |
| `go.go` | MOD: build tag `!avx,!sse` → `!avx,!sse,!arm64` | route arm64 to NEON | T20 |

## Re-sync notes

Same shape as the vecf32 vendor. On re-sync, re-apply the one-line `go.go`
build-tag edit + re-copy `go_arm64.go` + `add_arm64.s`.
