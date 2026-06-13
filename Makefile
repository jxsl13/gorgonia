# Developer tasks for gorgonia. Mirrors the CI pre-check gate (SPEC §V27/§V28,
# T24/T25) so issues can be seen and fixed locally.
#
#   make tools       install staticcheck + govulncheck
#   make fmt         gofmt -w (excludes vendored copies)
#   make lint        staticcheck (advisory — gorgonia has ~267 legacy issues)
#   make vuln        govulncheck
#   make test        go test
#   make pre-check   the exact CI gate (fails on any git diff)
#   make check       vet + lint + vuln + test

export GOFLAGS := -mod=mod

# Go files excluding vendored copies (kept verbatim, SPEC §V20).
GOFILES := $(shell find . -name '*.go' -not -path './internal/vendor/*')
# Library packages (CUDA/BLAS code is build-tag gated; examples/cmd skipped).
PKGS := $(shell go list ./... | grep -vE '/examples/|/cmd/')

.PHONY: tools fmt tidy vet lint vuln test pre-check check

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

fmt:
	gofmt -w $(GOFILES)

tidy:
	go mod tidy

vet:
	go vet $(PKGS)

# Advisory: prints all staticcheck findings so they can be fixed. The gorgonia
# library carries ~267 pre-existing issues; our new packages are clean.
lint:
	staticcheck $(PKGS)

vuln:
	govulncheck $(PKGS)

test:
	go test $(PKGS)

# Mirror of the CI pre-check gate: each mutating command must leave no diff.
pre-check:
	@unformatted="$$(gofmt -l $(GOFILES))"; \
	  if [ -n "$$unformatted" ]; then echo "gofmt needed:"; echo "$$unformatted"; exit 1; fi
	go generate ./...
	git diff --exit-code
	go mod tidy
	git diff --exit-code go.mod go.sum
	govulncheck $(PKGS)
	@echo "pre-check OK"

check: vet lint vuln test
