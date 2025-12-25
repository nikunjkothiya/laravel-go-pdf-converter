#!/bin/bash

# Exit on error
set -e

echo "Building Go binaries for distribution..."

# Ensure bin directory exists
mkdir -p bin

# Navigate to Go source
cd go-binary

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o ../bin/gopdfconv-linux-amd64 cmd/gopdfconv/main.go

# Build for Linux (arm64)
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 /usr/local/go/bin/go build -o ../bin/gopdfconv-linux-arm64 cmd/gopdfconv/main.go

# Build for macOS (amd64)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 /usr/local/go/bin/go build -o ../bin/gopdfconv-darwin-amd64 cmd/gopdfconv/main.go

# Build for macOS (arm64 - Apple Silicon)
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 /usr/local/go/bin/go build -o ../bin/gopdfconv-darwin-arm64 cmd/gopdfconv/main.go

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 /usr/local/go/bin/go build -o ../bin/gopdfconv-windows-amd64.exe cmd/gopdfconv/main.go

echo "Build complete! Binaries are in bin/"
ls -lh ../bin/
