# UPSTREAM provenance

- **Upstream module:** `github.com/gomlx/go-coreml`
- **Upstream repo:** https://github.com/gomlx/go-coreml
- **Vendored version:** `v0.0.0-20260301010621-8fdf6ad8655e` (alpha)
- **Vendored commit:** `8fdf6ad8655e`
- **Sync date:** 2026-06-13
- **License:** Apache-2.0 (CoreML protobuf defs BSD-3 per upstream) — `./LICENSE`
- **go.mod replace:** `replace github.com/gomlx/go-coreml => ./internal/vendor/go-coreml`

## Our modifications

| file | change | reason | SPEC ref |
|---|---|---|---|
| `runtime/computeunits_public.go` | added | re-export `ComputeUnits` + constants publicly (upstream keeps them in `internal/bridge`, unreachable externally) | T15, §C12, I.coreml |

`_test.go` files were dropped from the vendored copy (not needed for our build).

## Re-sync notes

Upstream's `WithComputeUnits` takes `internal/bridge.ComputeUnits`, which external
code cannot construct — hence the one added file. On re-sync, re-copy upstream
tree (minus tests) and re-add `runtime/computeunits_public.go`.
