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

# Pinned digests for reproducible builds (updated: 2025-04-23)
GOLANG_IMAGE := golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1
ALPINE_IMAGE := alpine:3.19@sha256:6baf43584bcb78f2e5847d1de515f23499913ac9f12bdf834811a3145eb11ca1

DOCKER_GO := docker run --rm -w /build -v $(shell pwd):/build $(GOLANG_IMAGE) go

# ============================================================================
# Targets
# ============================================================================

.PHONY: generate-mod
generate-mod: ## Generate go.mod and go.sum in Docker, copy to host
	@echo "Generating go.mod and go.sum in Docker container..."
	@docker run --rm -w /build -v $(shell pwd):/build \
		$(GOLANG_IMAGE) sh -c '\
		apk add --no-cache git; \
		if [ ! -f go.mod ]; then \
			go mod init github.com/glemsom/dkvmmanager; \
		fi; \
		go mod download; \
		go mod tidy'
	@chown $$(id -u):$$(id -g) go.mod go.sum 2>/dev/null || true
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
test: generate-mod ## Run go vet and all tests (COVER=1 for coverage, TEST_FLAGS for extra args)
	@echo "=== Running go vet and tests ==="
	@docker run --rm -w /build -v $(shell pwd):/build -e GOCACHE=/tmp/go-cache \
		--user $$(id -u):$$(id -g) \
		$(GOLANG_IMAGE) \
		sh -c '\
			echo "Running go vet..."; \
			go vet ./... > /tmp/vet.log 2>&1; \
			VET_EXIT=$$?; \
			cat /tmp/vet.log; \
			if [ "$$VET_EXIT" -ne 0 ]; then \
				echo "go vet failed"; \
				exit $$VET_EXIT; \
			fi; \
			echo ""; \
			echo "Running tests..."; \
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
