#!/bin/bash
set -e

echo "Building Go application..."
echo "Go version: $(go version)"
echo "Current directory: $(pwd)"
echo "Listing files:"
ls -la

echo "Building api/index.go..."
go build -o index api/index.go

echo "Build completed successfully!"
echo "Build output:"
ls -la
