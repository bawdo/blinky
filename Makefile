VERSION := $(shell git describe --tags --always --dirty)
COMMIT  := $(shell git rev-parse HEAD)
TAG     := $(shell git describe --exact-match --tags HEAD 2>/dev/null)
DIRTY   := $(shell test -n "$$(git status --porcelain)" && echo true || echo false)
LDFLAGS := -X github.com/bawdo/blinky/internal/version.Version=$(VERSION) \
           -X github.com/bawdo/blinky/internal/version.Commit=$(COMMIT) \
           -X github.com/bawdo/blinky/internal/version.Tag=$(TAG) \
           -X github.com/bawdo/blinky/internal/version.Dirty=$(DIRTY)

VULN_VERSION := v1.9.0

.DEFAULT_GOAL := help

build: ## Build the blinky binary into bin/ with version info embedded.
	go build -ldflags "$(LDFLAGS)" -o bin/blinky .

install: ## Install blinky into your Go bin directory with version info embedded.
	go install -ldflags "$(LDFLAGS)" .

test: ## Run the unit test suite.
	go test ./...

lint: ## Run golangci-lint over the codebase.
	golangci-lint run

vuln: ## Run govulncheck over the module, pinned to VULN_VERSION.
	go run golang.org/x/vuln/cmd/govulncheck@$(VULN_VERSION) ./...

integration: ## Run the integration tests (needs the integration build tag).
	go test -tags integration -v ./test/integration/...

pre-ci: ## Run the full set of local pre-CI quality checks before pushing.
	./scripts/pre-ci-check.sh

pre-ci-fix: ## Auto-fix gofmt issues, then run the pre-CI quality checks.
	./scripts/pre-ci-check.sh --fix gofmt

test-scripts: ## Run the shell script test suites.
	@for t in $$(ls scripts/*_test.sh | sort); do echo "==> $$t"; bash "$$t" || exit 1; done

help: ## List all make targets with a short description.
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: build install test test-scripts integration lint vuln pre-ci pre-ci-fix help
