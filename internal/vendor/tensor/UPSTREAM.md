# UPSTREAM provenance

- **Upstream module:** `gorgonia.org/tensor`
- **Upstream repo:** https://github.com/gorgonia/tensor
- **Vendored version:** `v0.9.24`
- **Sync date:** 2026-06-13
- **License:** Apache-2.0 (gorgonia org)
- **go.mod replace:** `replace gorgonia.org/tensor => ./internal/vendor/tensor`

## Our modifications

| file | change | reason | SPEC ref |
|---|---|---|---|
| `engine.go` | add `Pointer() unsafe.Pointer` to `Memory` (the doc already referenced it) | let callers reconstruct a typed pointer without the vet-flagged uintptr round-trip | T31, B8 |
| `array.go` | add `(*array).Pointer()` | implement the new method | T31 |
| `sparse.go` | add `(*CS).Pointer()` (delegates to array) | implement the new method | T31 |

Tests stripped from the vendored copy.
