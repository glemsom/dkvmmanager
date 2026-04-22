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
test: ## Run all tests and show summary only
	@echo "Running tests..."
	@docker run --rm -w /build -v $(shell pwd):/build golang:1.26-alpine sh -c '\
	go test ./... 2>&1 | tee /tmp/test.txt; \
	PASS=$$(awk "/^ok/{print 1}" /tmp/test.txt | wc -l); \
	FAIL=$$(awk "/^FAIL/{print 1}" /tmp/test.txt | wc -l); \
	echo ""; \
	echo "=== Test Summary ==="; \
	echo "Passed:  $$PASS"; \
	echo "Failed:  $$FAIL"; \
	echo "==================="; \
	if [ "$$FAIL" -gt 0 ]; then \
		awk "/^FAIL.*--- FAIL/{print}" /tmp/test.txt; \
	fi; \
	[ "$$FAIL" -eq 0 ]'

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
	@echo "╚═════════════════════════════════════════════════��═════════════╝"
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