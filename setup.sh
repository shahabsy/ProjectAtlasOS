#!/bin/bash
set -e

echo "Setting up Atlas Phase 0 environment..."

# Ensure Go is installed
if ! command -v go &> /dev/null; then
    echo " Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Get dependencies
go mod tidy

echo "Setup complete."
echo " Run: go build -o atlas.exe . "