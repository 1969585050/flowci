#!/bin/bash

set -e

ADDR=${FLOWCI_API_ADDR:-"localhost:3847"}
LOG_LEVEL=${FLOWCI_LOG_LEVEL:-"info"}

echo "=========================================="
echo "FlowCI Go API Server"
echo "=========================================="
echo "Address: $ADDR"
echo "Log Level: $LOG_LEVEL"
echo ""

cd "$(dirname "$0")/.."

export FLOWCI_API_ADDR
export FLOWCI_LOG_LEVEL

echo "Starting FlowCI API Server..."
go run cmd/server/main.go
