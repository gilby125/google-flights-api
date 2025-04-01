#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Running all Go tests..."

# Run all tests recursively, including integration tests.
# The -v flag provides verbose output.
go test ./... -v

echo "All tests completed."
