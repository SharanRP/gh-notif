#!/bin/bash

# Comprehensive End-to-End Test Runner for gh-notif
# This script runs all test suites and generates a comprehensive report

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-30m}
GITHUB_TOKEN=${GITHUB_TOKEN:-""}
PARALLEL_TESTS=${PARALLEL_TESTS:-4}
REPORT_DIR=${REPORT_DIR:-"test-reports"}
VERBOSE=${VERBOSE:-false}

# Test suites
SYSTEM_TESTS="tests/system"
E2E_TESTS="tests/e2e"
ACCEPTANCE_TESTS="tests/acceptance"
SECURITY_TESTS="tests/security"
DISTRIBUTION_TESTS="tests/distribution"

# Functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        error "Go is required but not installed"
    fi
    
    # Check Git
    if ! command -v git &> /dev/null; then
        error "Git is required but not installed"
    fi
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
        error "Please run this script from the gh-notif root directory"
    fi
    
    # Check GitHub token for tests that need it
    if [ -z "$GITHUB_TOKEN" ]; then
        warn "GITHUB_TOKEN not set - some tests will be skipped"
    fi
    
    success "Prerequisites check passed"
}

# Setup test environment
setup_test_environment() {
    log "Setting up test environment..."
    
    # Create report directory
    mkdir -p "$REPORT_DIR"
    
    # Clean previous test artifacts
    rm -rf "$REPORT_DIR"/*
    
    # Build the binary for testing
    log "Building gh-notif for testing..."
    go build -o gh-notif-test ./main.go
    
    # Set environment variables
    export GH_NOTIF_TEST_BINARY="$(pwd)/gh-notif-test"
    export GH_NOTIF_TEST_MODE="true"
    
    success "Test environment setup complete"
}

# Run unit tests
run_unit_tests() {
    log "Running unit tests..."
    
    local report_file="$REPORT_DIR/unit-tests.xml"
    local coverage_file="$REPORT_DIR/coverage.out"
    
    if $VERBOSE; then
        go test -v -race -coverprofile="$coverage_file" -timeout="$TEST_TIMEOUT" ./... | tee "$REPORT_DIR/unit-tests.log"
    else
        go test -race -coverprofile="$coverage_file" -timeout="$TEST_TIMEOUT" ./... > "$REPORT_DIR/unit-tests.log" 2>&1
    fi
    
    local exit_code=$?
    
    # Generate coverage report
    if [ -f "$coverage_file" ]; then
        go tool cover -html="$coverage_file" -o "$REPORT_DIR/coverage.html"
        go tool cover -func="$coverage_file" > "$REPORT_DIR/coverage.txt"
        
        # Extract coverage percentage
        local coverage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}')
        log "Code coverage: $coverage"
    fi
    
    if [ $exit_code -eq 0 ]; then
        success "Unit tests passed"
    else
        error "Unit tests failed"
    fi
    
    return $exit_code
}

# Run system tests
run_system_tests() {
    log "Running system tests..."
    
    local report_file="$REPORT_DIR/system-tests.xml"
    
    if [ -d "$SYSTEM_TESTS" ]; then
        if $VERBOSE; then
            go test -v -timeout="$TEST_TIMEOUT" "./$SYSTEM_TESTS/..." | tee "$REPORT_DIR/system-tests.log"
        else
            go test -timeout="$TEST_TIMEOUT" "./$SYSTEM_TESTS/..." > "$REPORT_DIR/system-tests.log" 2>&1
        fi
        
        local exit_code=$?
        
        if [ $exit_code -eq 0 ]; then
            success "System tests passed"
        else
            warn "System tests failed or skipped"
        fi
        
        return $exit_code
    else
        warn "System tests directory not found: $SYSTEM_TESTS"
        return 0
    fi
}

# Run end-to-end tests
run_e2e_tests() {
    log "Running end-to-end tests..."
    
    local report_file="$REPORT_DIR/e2e-tests.xml"
    
    if [ -d "$E2E_TESTS" ]; then
        if $VERBOSE; then
            go test -v -timeout="$TEST_TIMEOUT" "./$E2E_TESTS/..." | tee "$REPORT_DIR/e2e-tests.log"
        else
            go test -timeout="$TEST_TIMEOUT" "./$E2E_TESTS/..." > "$REPORT_DIR/e2e-tests.log" 2>&1
        fi
        
        local exit_code=$?
        
        if [ $exit_code -eq 0 ]; then
            success "End-to-end tests passed"
        else
            warn "End-to-end tests failed or skipped"
        fi
        
        return $exit_code
    else
        warn "E2E tests directory not found: $E2E_TESTS"
        return 0
    fi
}

# Run security tests
run_security_tests() {
    log "Running security tests..."
    
    local report_file="$REPORT_DIR/security-tests.xml"
    
    if [ -d "$SECURITY_TESTS" ]; then
        if $VERBOSE; then
            go test -v -timeout="$TEST_TIMEOUT" "./$SECURITY_TESTS/..." | tee "$REPORT_DIR/security-tests.log"
        else
            go test -timeout="$TEST_TIMEOUT" "./$SECURITY_TESTS/..." > "$REPORT_DIR/security-tests.log" 2>&1
        fi
        
        local exit_code=$?
        
        if [ $exit_code -eq 0 ]; then
            success "Security tests passed"
        else
            warn "Security tests failed or skipped"
        fi
        
        return $exit_code
    else
        warn "Security tests directory not found: $SECURITY_TESTS"
        return 0
    fi
}

# Run distribution tests
run_distribution_tests() {
    log "Running distribution tests..."
    
    local report_file="$REPORT_DIR/distribution-tests.xml"
    
    if [ -d "$DISTRIBUTION_TESTS" ]; then
        if $VERBOSE; then
            go test -v -timeout="$TEST_TIMEOUT" "./$DISTRIBUTION_TESTS/..." | tee "$REPORT_DIR/distribution-tests.log"
        else
            go test -timeout="$TEST_TIMEOUT" "./$DISTRIBUTION_TESTS/..." > "$REPORT_DIR/distribution-tests.log" 2>&1
        fi
        
        local exit_code=$?
        
        if [ $exit_code -eq 0 ]; then
            success "Distribution tests passed"
        else
            warn "Distribution tests failed or skipped"
        fi
        
        return $exit_code
    else
        warn "Distribution tests directory not found: $DISTRIBUTION_TESTS"
        return 0
    fi
}

# Run performance benchmarks
run_performance_benchmarks() {
    log "Running performance benchmarks..."
    
    local bench_file="$REPORT_DIR/benchmarks.txt"
    
    go test -bench=. -benchmem -timeout="$TEST_TIMEOUT" ./... > "$bench_file" 2>&1
    
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        success "Performance benchmarks completed"
    else
        warn "Performance benchmarks failed"
    fi
    
    return $exit_code
}

# Run static analysis
run_static_analysis() {
    log "Running static analysis..."
    
    # golangci-lint
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run --out-format=json > "$REPORT_DIR/lint-report.json" 2>&1
        golangci-lint run > "$REPORT_DIR/lint-report.txt" 2>&1
        
        if [ $? -eq 0 ]; then
            success "Linting passed"
        else
            warn "Linting found issues"
        fi
    else
        warn "golangci-lint not found, skipping linting"
    fi
    
    # gosec
    if command -v gosec &> /dev/null; then
        gosec -fmt json -out "$REPORT_DIR/security-scan.json" ./...
        gosec -fmt text -out "$REPORT_DIR/security-scan.txt" ./...
        
        if [ $? -eq 0 ]; then
            success "Security scan passed"
        else
            warn "Security scan found issues"
        fi
    else
        warn "gosec not found, skipping security scan"
    fi
    
    # govulncheck
    if command -v govulncheck &> /dev/null; then
        govulncheck ./... > "$REPORT_DIR/vulnerability-check.txt" 2>&1
        
        if [ $? -eq 0 ]; then
            success "Vulnerability check passed"
        else
            warn "Vulnerability check found issues"
        fi
    else
        warn "govulncheck not found, skipping vulnerability check"
    fi
}

# Generate test report
generate_test_report() {
    log "Generating test report..."
    
    local report_file="$REPORT_DIR/test-summary.md"
    
    cat > "$report_file" << EOF
# gh-notif End-to-End Test Report

**Generated:** $(date)
**Platform:** $(uname -s) $(uname -m)
**Go Version:** $(go version)

## Test Results Summary

EOF
    
    # Add test results
    for log_file in "$REPORT_DIR"/*.log; do
        if [ -f "$log_file" ]; then
            local test_name=$(basename "$log_file" .log)
            echo "### $test_name" >> "$report_file"
            
            if grep -q "PASS" "$log_file"; then
                echo "✅ **PASSED**" >> "$report_file"
            elif grep -q "FAIL" "$log_file"; then
                echo "❌ **FAILED**" >> "$report_file"
            else
                echo "⚠️ **SKIPPED**" >> "$report_file"
            fi
            
            echo "" >> "$report_file"
        fi
    done
    
    # Add coverage information
    if [ -f "$REPORT_DIR/coverage.txt" ]; then
        echo "## Code Coverage" >> "$report_file"
        echo "\`\`\`" >> "$report_file"
        cat "$REPORT_DIR/coverage.txt" >> "$report_file"
        echo "\`\`\`" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # Add benchmark results
    if [ -f "$REPORT_DIR/benchmarks.txt" ]; then
        echo "## Performance Benchmarks" >> "$report_file"
        echo "\`\`\`" >> "$report_file"
        grep "Benchmark" "$REPORT_DIR/benchmarks.txt" >> "$report_file"
        echo "\`\`\`" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    success "Test report generated: $report_file"
}

# Cleanup
cleanup() {
    log "Cleaning up..."
    
    # Remove test binary
    rm -f gh-notif-test
    
    # Unset environment variables
    unset GH_NOTIF_TEST_BINARY
    unset GH_NOTIF_TEST_MODE
    
    success "Cleanup complete"
}

# Main execution
main() {
    log "Starting comprehensive end-to-end testing for gh-notif"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --verbose|-v)
                VERBOSE=true
                shift
                ;;
            --timeout|-t)
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            --report-dir|-r)
                REPORT_DIR="$2"
                shift 2
                ;;
            --help|-h)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --verbose, -v          Enable verbose output"
                echo "  --timeout, -t TIMEOUT  Set test timeout (default: 30m)"
                echo "  --report-dir, -r DIR   Set report directory (default: test-reports)"
                echo "  --help, -h             Show this help message"
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    # Run test phases
    check_prerequisites
    setup_test_environment
    
    # Track test results
    local unit_result=0
    local system_result=0
    local e2e_result=0
    local security_result=0
    local distribution_result=0
    
    # Run all test suites
    run_unit_tests || unit_result=$?
    run_system_tests || system_result=$?
    run_e2e_tests || e2e_result=$?
    run_security_tests || security_result=$?
    run_distribution_tests || distribution_result=$?
    
    # Run additional checks
    run_performance_benchmarks
    run_static_analysis
    
    # Generate report
    generate_test_report
    
    # Cleanup
    cleanup
    
    # Final summary
    log "Test execution complete"
    log "Results:"
    log "  Unit tests: $([ $unit_result -eq 0 ] && echo "PASS" || echo "FAIL")"
    log "  System tests: $([ $system_result -eq 0 ] && echo "PASS" || echo "FAIL")"
    log "  E2E tests: $([ $e2e_result -eq 0 ] && echo "PASS" || echo "FAIL")"
    log "  Security tests: $([ $security_result -eq 0 ] && echo "PASS" || echo "FAIL")"
    log "  Distribution tests: $([ $distribution_result -eq 0 ] && echo "PASS" || echo "FAIL")"
    
    # Exit with error if any critical tests failed
    if [ $unit_result -ne 0 ]; then
        error "Unit tests failed - this is a critical failure"
    fi
    
    if [ $system_result -ne 0 ] || [ $e2e_result -ne 0 ] || [ $security_result -ne 0 ]; then
        warn "Some test suites failed - review the reports"
        exit 1
    fi
    
    success "All tests completed successfully!"
}

# Run main function
main "$@"
