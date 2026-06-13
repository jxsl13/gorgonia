# internal/vendor — modified dependency copies

Convention for vendoring dependencies that need **our** modifications.
Codifies SPEC.md §C12 + §V20–V23.

## When to vendor

Vendor a dependency **only** when this repo must modify it (SPEC §V22). A dep
that works unmodified stays a normal external `require` — do **not** vendor it
(e.g. `go4.org/unsafe/assume-no-moving-gc`, SPEC §C6).

## Layout

```
internal/vendor/<name>/
  UPSTREAM.md      # provenance — see UPSTREAM.template.md
  LICENSE          # upstream license, verbatim (SPEC §V20)
  <source...>      # verbatim copy of upstream latest release
  <name>_*.go|.s   # OUR additions, in separate files where possible (§V20)
```

`<name>` is the last path element of the upstream module
(e.g. `vecf32` for `gorgonia.org/vecf32`).

## Rules

1. **Start from upstream LATEST release** — copy, then layer our diff on top.
   Not a rewrite (SPEC §V21). The baseline must stay diffable against pristine
   upstream so re-syncs are mechanical.
2. **Isolate our changes.** Put additions in new files (`*_arm64.s`,
   `*_neon.go`, …) rather than editing upstream files in place, wherever the
   upstream structure allows. If an in-place edit is unavoidable, mark it with a
   `// MOD(jxsl13):` comment.
3. **Preserve LICENSE verbatim** and record provenance in `UPSTREAM.md`
   (module path, exact version/commit, sync date).
4. **Wire via `replace`** in the root `go.mod`:
   ```
   replace gorgonia.org/vecf32 => ./internal/vendor/vecf32
   ```
   This makes transitive importers (e.g. `tensor` → `vecf32`) resolve to our
   copy too (SPEC §V23).
5. The vendored module keeps the **original import path / package name** so the
   `replace` is transparent. Add a `go.mod` inside the vendored dir with the
   original module path.

## Re-sync procedure

1. Diff `internal/vendor/<name>/` against the new pristine upstream release.
2. Re-apply our isolated files / `MOD(jxsl13):` hunks.
3. Update `UPSTREAM.md` (version/commit, sync date).
4. `go mod verify` + build + parity tests green.

## Verification

- `go build ./...` + `go mod verify` green with `replace` active.
- Pristine build (our mods reverted) parity-tested vs modified (SPEC §V23).
