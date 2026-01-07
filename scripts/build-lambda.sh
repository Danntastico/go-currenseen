#!/bin/bash
# Build script for Lambda function
# This script builds the Go binary for AWS Lambda

set -e

echo "Building Lambda function..."

# Set build variables
BINARY_NAME="bootstrap"
LAMBDA_DIR="cmd/lambda"
BUILD_DIR=".aws-sam/build/ExchangeRateFunction"

# Create build directory
mkdir -p "$BUILD_DIR"

# Build Go binary for Linux/AMD64 (Lambda runtime)
echo "Compiling Go binary for Linux/AMD64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o "$BUILD_DIR/$BINARY_NAME" \
    "./$LAMBDA_DIR"

# Make binary executable
chmod +x "$BUILD_DIR/$BINARY_NAME"

# Verify binary
if [ -f "$BUILD_DIR/$BINARY_NAME" ]; then
    echo "✓ Build successful: $BUILD_DIR/$BINARY_NAME"
    ls -lh "$BUILD_DIR/$BINARY_NAME"
else
    echo "✗ Build failed: binary not found"
    exit 1
fi

echo "Lambda function build complete!"

