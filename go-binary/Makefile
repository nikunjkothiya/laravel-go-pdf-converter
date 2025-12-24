.PHONY: all build test clean install

# Variables
BINARY_NAME=gopdfconv
BUILD_DIR=../bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build for current platform
build: deps
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gopdfconv

# Build for all platforms
build-all: deps
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/gopdfconv
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/gopdfconv
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/gopdfconv
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/gopdfconv
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/gopdfconv

# Run tests
test:
	go test -v -race -cover ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install locally
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build for current platform"
	@echo "  build-all  - Build for all platforms (Linux, macOS, Windows)"
	@echo "  test       - Run tests"
	@echo "  bench      - Run benchmarks"
	@echo "  coverage   - Generate coverage report"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install to /usr/local/bin"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  deps       - Download dependencies"
