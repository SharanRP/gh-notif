#!/bin/bash

# Test coverage script for bash

# Create coverage directory if it doesn't exist
COVERAGE_DIR="coverage"
mkdir -p "$COVERAGE_DIR"

# Run tests with coverage
echo "Running tests with coverage..."
go test ./... -coverprofile="$COVERAGE_DIR/coverage.out" -covermode=atomic

# Generate HTML coverage report
echo "Generating HTML coverage report..."
go tool cover -html="$COVERAGE_DIR/coverage.out" -o="$COVERAGE_DIR/coverage.html"

# Calculate coverage percentage
echo "Coverage summary:"
go tool cover -func="$COVERAGE_DIR/coverage.out"

# Open the coverage report in the default browser
echo "Opening coverage report in browser..."
case "$(uname -s)" in
    Darwin*)    open "$COVERAGE_DIR/coverage.html" ;;
    Linux*)     xdg-open "$COVERAGE_DIR/coverage.html" ;;
    *)          echo "Please open $COVERAGE_DIR/coverage.html in your browser" ;;
esac
