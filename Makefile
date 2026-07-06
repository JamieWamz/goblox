.PHONY: help build test coverage run clean install fmt lint

BINARY_NAME=goblox
BUILD_DIR=bin

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goblox
	@echo "Done! Binary at $(BUILD_DIR)/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

run:
	@go run ./cmd/goblox/main.go

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Done!"

fmt:
	@go fmt ./...

lint:
	@golangci-lint run ./...

install:
	@go install ./cmd/goblox
