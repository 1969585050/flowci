#!/bin/bash

set -e

VERSION=${1:-"0.1.0"}
OUTPUT_DIR=${2:-"./release"}

echo "=========================================="
echo "FlowCI Build Script"
echo "Version: $VERSION"
echo "=========================================="

mkdir -p "$OUTPUT_DIR"

echo ""
echo "[1/4] Building Go API server..."
cd go
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$OUTPUT_DIR/flowci-api" cmd/server/main.go
echo "      Go API server built: $OUTPUT_DIR/flowci-api"

cd ..

echo ""
echo "[2/4] Building frontend..."
cd src
npm run build
mkdir -p "$OUTPUT_DIR/web"
cp -r dist/* "$OUTPUT_DIR/web/"
cd ..
echo "      Frontend built: $OUTPUT_DIR/web/"

echo ""
echo "[3/4] Checking Tauri..."
if command -v cargo &> /dev/null; then
    echo "      Tauri available, skipping desktop build (use 'make build-tauri' for desktop)"
else
    echo "      Tauri not available, skipping desktop build"
fi

echo ""
echo "[4/4] Creating release package..."
cd "$OUTPUT_DIR"
tar -czvf "flowci-$VERSION.tar.gz" flowci-api web
cd ..

echo ""
echo "=========================================="
echo "Build complete!"
echo "=========================================="
echo "Release files:"
echo "  - $OUTPUT_DIR/flowci-api"
echo "  - $OUTPUT_DIR/web/"
echo "  - $OUTPUT_DIR/flowci-$VERSION.tar.gz"
echo ""
