.PHONY: help build test coverage run clean install fmt fmt-check lint check

BINARY_NAME := goblox
BUILD_DIR := bin

help: ## Show available targets
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the CLI binary
	@mkdir -p $(BUILD_DIR)
	go build -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goblox

test: ## Run race-enabled tests with package coverage
	go test -race -cover ./...

coverage: ## Generate HTML and text coverage reports
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

run: ## Run the CLI from source
	go run ./cmd/goblox

clean: ## Remove generated build and coverage artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

fmt: ## Format Go source files
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

fmt-check: ## Verify Go source formatting
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))"

lint: fmt-check ## Run static analysis
	go vet ./...

check: lint test build ## Run all required quality checks

install: ## Install goblox with go install
	go install ./cmd/goblox
