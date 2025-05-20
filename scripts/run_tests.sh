#!/bin/bash
# Bash script to run tests with coverage reporting

# Create a directory for coverage reports if it doesn't exist
COVERAGE_DIR="coverage"
mkdir -p "$COVERAGE_DIR"

# Clean up any previous coverage files
rm -f "$COVERAGE_DIR"/*

# Run tests with coverage
echo "Running tests with coverage..."
go test ./... -coverprofile="$COVERAGE_DIR/coverage.out" -covermode=atomic

# Check if the tests passed
if [ $? -ne 0 ]; then
    echo "Tests failed with exit code $?"
    exit 1
fi

# Generate HTML coverage report
echo "Generating HTML coverage report..."
go tool cover -html="$COVERAGE_DIR/coverage.out" -o="$COVERAGE_DIR/coverage.html"

# Generate coverage summary
echo "Generating coverage summary..."
go tool cover -func="$COVERAGE_DIR/coverage.out" | tee "$COVERAGE_DIR/coverage_summary.txt"

# Extract the total coverage percentage
TOTAL_COVERAGE=$(grep "total:" "$COVERAGE_DIR/coverage_summary.txt")
echo -e "\033[0;32mCoverage: $TOTAL_COVERAGE\033[0m"

# Open the coverage report in the default browser
echo "Opening coverage report in browser..."
case "$(uname -s)" in
    Darwin*)    open "$COVERAGE_DIR/coverage.html" ;;
    Linux*)     xdg-open "$COVERAGE_DIR/coverage.html" ;;
    *)          echo "Please open $COVERAGE_DIR/coverage.html in your browser" ;;
esac

echo -e "\033[0;32mTest run completed successfully!\033[0m"
