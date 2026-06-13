# UPSTREAM provenance

Copy this file to `internal/vendor/<name>/UPSTREAM.md` and fill it in.

- **Upstream module:** `<module path, e.g. gorgonia.org/vecf32>`
- **Upstream repo:** `<URL>`
- **Vendored version:** `<semver tag or pseudo-version>`
- **Vendored commit:** `<full git SHA>`
- **Sync date:** `<YYYY-MM-DD>`
- **License:** `<SPDX id, e.g. BSD-3-Clause>` (copy preserved in `./LICENSE`)
- **go.mod replace:** `replace <module path> => ./internal/vendor/<name>`

## Our modifications

| file | change | reason | SPEC ref |
|---|---|---|---|
| `<name>_arm64.s` | added | ARM64 NEON impl | T20, §V11 |
| ... | ... | ... | ... |

## Re-sync notes

`<anything tricky about re-applying our diff to a newer upstream release>`
