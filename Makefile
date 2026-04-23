# DKVM Manager Makefile
# Containerized builds using Docker

# ============================================================================
# Version and Build Configuration
# ============================================================================

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

OUTPUT  := dkvmmanager

# ============================================================================
# Docker Configuration
# ============================================================================

DOCKER_GO := docker run --rm -w /build -v $(shell pwd):/build golang:1.26-alpine go

# ============================================================================
# Targets
# ============================================================================

.PHONY: update-golden
update-golden: ## Update golden test files (UPDATE_GOLDEN=1)
	@echo "Updating golden files..."
	@docker run --rm -w /build -v $(shell pwd):/build --user $$(id -u):$$(id -g) \
		-e UPDATE_GOLDEN=1 \
		-e GOCACHE=/tmp/go-cache \
		golang:1.26-alpine go test -v ./internal/tui/models/...
	@echo "Golden files updated."

.PHONY: generate-mod
generate-mod: ## Generate go.mod and go.sum in Docker, copy to host
	@echo "Generating go.mod and go.sum in Docker container..."
	@docker run --rm -w /build -v $(shell pwd):/build --user $$(id -u):$$(id -g) \
		-e GOCACHE=/tmp/go-cache \
		golang:1.26-alpine sh -c '\
		if [ ! -f go.mod ]; then \
			go mod init github.com/glemsom/dkvmmanager; \
		fi; \
		go mod tidy'
	@echo "Done: go.mod and go.sum generated."

.PHONY: coverage
coverage: ## Run tests with coverage (HTML report in coverage.html)
	@echo "Running tests with coverage..."
	@docker run --rm -w /build -v $(shell pwd):/build \
		golang:1.26-alpine sh -c '\
			go test -v -coverprofile=/build/coverage.out ./... 2>&1 | tee /tmp/coverage.log; \
			if [ -f /build/coverage.out ]; then \
				go tool cover -html=/build/coverage.out -o /build/coverage.html; \
			fi; \
		' || true
	@echo ""; \
	if [ -f coverage.out ]; then \
		go tool cover -func=coverage.out | tail -1; \
		echo "\nHTML report: file://$(shell pwd)/coverage.html"; \
	else \
		echo "Error: coverage.out not generated on host"; \
	fi

.PHONY: test-short
test-short: ## Run tests with -short flag (skips integration tests)
	@$(MAKE) test TEST_FLAGS="-short"

.PHONY: build
build: ## Build application in Docker using go.mod/go.sum from host
	@echo "Building $(OUTPUT)..."
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE="$(DATE)" \
		--target builder \
		-t dkvmmanager:build .
	@docker run --rm -v $(shell pwd):/output --entrypoint cp dkvmmanager:build /build/$(OUTPUT) /output/
	@docker rmi dkvmmanager:build 2>/dev/null || true
	@echo "Built: $(OUTPUT)"

.PHONY: test
# Go test flags can be passed via TEST_FLAGS (e.g. make test TEST_FLAGS="-v -run TestName")
# Use COVER=1 to enable coverage (e.g. make test COVER=1)
test: ## Run all tests (COVER=1 for coverage, TEST_FLAGS for extra args)
	@echo "Running tests..."
	@docker run --rm -w /build -v $(shell pwd):/build -e GOCACHE=/tmp/go-cache \
		golang:1.26-alpine \
		sh -c '\
			FLAGS="$(TEST_FLAGS)"; \
			if [ "$(COVER)" = "1" ]; then FLAGS="$$FLAGS -cover"; fi; \
			go test $$FLAGS ./... > /tmp/test.log 2>&1; \
			EXIT=$$?; \
			cat /tmp/test.log; \
			PASS=$$(grep -c "^ok" /tmp/test.log || true); \
			FAIL=$$(grep -c "^FAIL" /tmp/test.log || true); \
			echo ""; \
			echo "=== Test Summary ==="; \
			echo "Passed:  $$PASS"; \
			echo "Failed:  $$FAIL"; \
			echo "==================="; \
			if [ "$$FAIL" -gt 0 ]; then \
				echo ""; \
				grep "^FAIL" /tmp/test.log; \
			fi; \
			exit $$EXIT; \
		'

.PHONY: run-dry
run-dry: build ## Build and run in dry-run mode (shows QEMU args without launching)
	@./$(OUTPUT) -dry-run

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(OUTPUT)

.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════╗"
	@echo "║              DKVM Manager - Build Targets                     ║"
	@echo "╚═══════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>    Run a build target"
	@echo "  make help       Show this help message"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-18s %s\n", $$1, $$2}'
	@echo ""
	@echo "Run 'make' without a target to build."
