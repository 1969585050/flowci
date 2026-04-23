#!/bin/bash

set -e

echo "=========================================="
echo "FlowCI Development Environment Setup"
echo "=========================================="

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "Error: $1 is required but not installed."
        exit 1
    fi
}

echo "Checking prerequisites..."
check_command git
check_command go
check_command node
check_command cargo

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
NODE_VERSION=$(node --version | grep -oP 'v\K[0-9]+\.[0-9]+')

echo "Detected versions:"
echo "  Go: $GO_VERSION"
echo "  Node: $NODE_VERSION"

if [[ $(echo "$GO_VERSION < 1.26" | bc) -eq 1 ]]; then
    echo "Warning: Go 1.26+ recommended"
fi

echo ""
echo "Setting up Go modules..."
cd go
go mod download
go mod tidy
cd ..

echo ""
echo "Setting up frontend..."
cd src
npm install
cd ..

echo ""
echo "Checking Rust toolchain..."
cargo --version
rustc --version

echo ""
echo "=========================================="
echo "Setup complete!"
echo "=========================================="
echo ""
echo "To start development:"
echo "  1. Start Go API server:"
echo "     cd go && go run cmd/server/main.go"
echo ""
echo "  2. Start frontend (in another terminal):"
echo "     cd src && npm run dev"
echo ""
echo "  3. Start Tauri (in another terminal):"
echo "     cd src-tauri && cargo tauri dev"
echo ""
