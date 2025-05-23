# Comprehensive Testing Guide for gh-notif

This guide provides detailed instructions for running the complete end-to-end testing suite for gh-notif before release.

## Overview

The testing suite consists of multiple layers designed to verify both functional and non-functional requirements:

1. **System Tests** - Installation methods and cross-platform compatibility
2. **User Acceptance Tests** - Complete workflow validation
3. **Distribution Tests** - Package format verification
4. **Security Audit** - Security vulnerability assessment
5. **Documentation Tests** - Documentation accuracy verification
6. **Performance Tests** - Performance benchmark validation

## Prerequisites

### Required Tools

- **Go 1.20+** - For building and running tests
- **Git** - For version control operations
- **Docker** (optional) - For container testing
- **Platform-specific package managers** (optional):
  - **macOS**: Homebrew
  - **Windows**: Scoop
  - **Linux**: Snap, dpkg, rpm

### Environment Setup

```bash
# Set GitHub token for API tests
export GITHUB_TOKEN="your_github_token_here"

# Optional: Set test configuration
export TEST_TIMEOUT="30m"
export PARALLEL_TESTS="4"
export REPORT_DIR="test-reports"
```

### Test Data Requirements

- GitHub account with notifications
- Access to test repositories
- Valid GitHub personal access token with scopes:
  - `notifications`
  - `repo:status`
  - `read:user`

## Running Tests

### Quick Start

```bash
# Run all tests (Linux/macOS)
./scripts/run-e2e-tests.sh

# Run all tests (Windows)
.\scripts\run-e2e-tests.ps1
```

### Individual Test Suites

#### 1. System Tests

Tests installation methods and cross-platform behavior:

```bash
# Run system tests
go test -v ./tests/system/...

# Test specific installation method
go test -v ./tests/system/ -run TestInstallationMethods/Binary

# Test cross-platform compatibility
go test -v ./tests/system/ -run TestCrossPlatformCompatibility
```

**Coverage:**
- Binary installation from releases
- Go install method
- Docker container usage
- Platform-specific package managers
- Cross-platform path handling
- File permission verification

#### 2. Workflow Tests

Tests complete user workflows:

```bash
# Run workflow tests
go test -v ./tests/system/ -run TestCompleteWorkflow

# Test specific workflow components
go test -v ./tests/system/ -run TestCompleteWorkflow/Authentication
go test -v ./tests/system/ -run TestCompleteWorkflow/Configuration
```

**Coverage:**
- Authentication flow
- Configuration management
- Notification listing and filtering
- Grouping and search functionality
- Action execution (read, open, subscribe)
- Export functionality

#### 3. Performance Tests

Tests performance against benchmarks:

```bash
# Run performance tests
go test -v ./tests/system/ -run TestPerformanceBenchmarks

# Run specific performance tests
go test -v ./tests/system/ -run TestPerformanceBenchmarks/StartupPerformance
go test -v ./tests/system/ -run TestResourceUsage
```

**Performance Criteria:**
- Startup time: < 2 seconds
- List response: < 5 seconds
- Filter response: < 3 seconds
- Search response: < 4 seconds
- Memory usage: < 100MB
- CPU usage: < 80%

#### 4. Distribution Tests

Tests package distribution methods:

```bash
# Run distribution tests
go test -v ./tests/distribution/...

# Test specific package format
go test -v ./tests/distribution/ -run TestPackageDistribution/Docker
go test -v ./tests/distribution/ -run TestPackageDistribution/Homebrew
```

**Coverage:**
- Docker image functionality
- Homebrew formula installation
- Scoop manifest installation
- Snap package installation
- DEB/RPM package installation
- Update mechanism verification

#### 5. Security Tests

Tests security aspects:

```bash
# Run security tests
go test -v ./tests/security/...

# Test specific security aspects
go test -v ./tests/security/ -run TestCredentialStorageSecurity
go test -v ./tests/security/ -run TestAuthenticationFlowSecurity
go test -v ./tests/security/ -run TestInputValidationSecurity
```

**Security Coverage:**
- Credential storage security
- Authentication flow security
- Network communication security
- Input validation and injection prevention
- File permission verification
- Memory security

#### 6. Documentation Tests

Tests documentation accuracy:

```bash
# Run documentation tests
go test -v ./tests/e2e/...

# Test specific documentation aspects
go test -v ./tests/e2e/ -run TestDocumentationAccuracy/READMEExamples
go test -v ./tests/e2e/ -run TestDocumentationAccuracy/HelpTextAccuracy
```

**Coverage:**
- README example verification
- Help text accuracy
- Command example validation
- Configuration documentation
- Man page generation

## User Acceptance Testing

### Manual Testing Checklist

Use the comprehensive checklist in `tests/acceptance/user_acceptance_test.md`:

1. **First-Time User Experience**
   - [ ] Setup wizard functionality
   - [ ] Authentication process
   - [ ] Tutorial completion
   - [ ] Initial notification listing

2. **Core Functionality**
   - [ ] Notification listing with various filters
   - [ ] Grouping by different criteria
   - [ ] Search functionality
   - [ ] Action execution

3. **Configuration Management**
   - [ ] Setting and getting configuration values
   - [ ] Configuration validation
   - [ ] Reset functionality

4. **Output Formats and Export**
   - [ ] JSON, CSV, and table formats
   - [ ] File export functionality
   - [ ] Large dataset handling

5. **Error Handling**
   - [ ] Invalid command handling
   - [ ] Network error scenarios
   - [ ] Permission error handling

6. **Performance and Usability**
   - [ ] Response time validation
   - [ ] Help system usability
   - [ ] Tab completion (where available)

### Test Scenarios

#### Scenario 1: New User Setup

```bash
# Clean environment
rm -rf ~/.gh-notif*

# First run
gh-notif firstrun

# Expected: Welcome message and setup wizard
# Validation: User can complete setup without errors
```

#### Scenario 2: Daily Usage Workflow

```bash
# Authenticate
gh-notif auth login

# List notifications
gh-notif list --limit 10

# Filter unread notifications
gh-notif list --filter "is:unread"

# Mark notification as read
gh-notif read [NOTIFICATION_ID]

# Expected: All operations complete successfully
# Validation: State changes are reflected correctly
```

#### Scenario 3: Advanced Features

```bash
# Complex filtering
gh-notif list --filter "is:unread AND type:PullRequest"

# Grouping
gh-notif group --by repository

# Search
gh-notif search "bug fix"

# Export
gh-notif list --format json --output notifications.json

# Expected: Advanced features work as documented
# Validation: Output matches expected format and content
```

## Test Reports

### Automated Report Generation

The test runner generates comprehensive reports:

```
test-reports/
├── test-summary.md          # Overall test summary
├── unit-tests.log           # Unit test results
├── system-tests.log         # System test results
├── e2e-tests.log           # E2E test results
├── security-tests.log      # Security test results
├── distribution-tests.log  # Distribution test results
├── coverage.html           # Code coverage report
├── coverage.txt            # Coverage summary
├── benchmarks.txt          # Performance benchmarks
├── lint-report.txt         # Linting results
├── security-scan.txt       # Security scan results
└── vulnerability-check.txt # Vulnerability check results
```

### Report Analysis

#### Coverage Analysis

```bash
# View coverage summary
cat test-reports/coverage.txt

# Open detailed coverage report
open test-reports/coverage.html
```

**Coverage Targets:**
- Overall coverage: > 80%
- Critical paths: > 95%
- New code: 100%

#### Performance Analysis

```bash
# View benchmark results
grep "Benchmark" test-reports/benchmarks.txt

# Check performance regressions
# Compare with previous benchmark results
```

#### Security Analysis

```bash
# Review security scan results
cat test-reports/security-scan.txt

# Check vulnerability report
cat test-reports/vulnerability-check.txt
```

## Continuous Integration

### GitHub Actions Integration

The testing suite integrates with GitHub Actions:

```yaml
# .github/workflows/e2e-tests.yml
name: End-to-End Tests
on: [push, pull_request]

jobs:
  e2e-tests:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.20'
    
    - name: Run E2E Tests
      run: ./scripts/run-e2e-tests.sh
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Upload Test Reports
      uses: actions/upload-artifact@v3
      with:
        name: test-reports-${{ matrix.os }}
        path: test-reports/
```

### Quality Gates

Tests must pass these quality gates:

1. **Unit Tests**: 100% pass rate
2. **Code Coverage**: > 80% overall
3. **Security Scan**: No high-severity issues
4. **Performance**: Within benchmark limits
5. **Documentation**: All examples work

## Troubleshooting

### Common Issues

#### Authentication Failures

```bash
# Issue: Tests fail with authentication errors
# Solution: Verify GitHub token
echo $GITHUB_TOKEN | gh auth login --with-token
gh auth status
```

#### Permission Errors

```bash
# Issue: File permission tests fail
# Solution: Check file system permissions
ls -la ~/.gh-notif*
```

#### Network Timeouts

```bash
# Issue: Tests timeout on network operations
# Solution: Increase timeout or check connectivity
export TEST_TIMEOUT="60m"
curl -I https://api.github.com
```

#### Package Manager Issues

```bash
# Issue: Package installation tests fail
# Solution: Verify package manager availability
which brew || which scoop || which snap
```

### Debug Mode

Enable verbose output for debugging:

```bash
# Linux/macOS
./scripts/run-e2e-tests.sh --verbose

# Windows
.\scripts\run-e2e-tests.ps1 -Verbose
```

### Test Isolation

Run tests in isolation to debug specific issues:

```bash
# Run single test
go test -v ./tests/system/ -run TestSpecificFunction

# Run with race detection
go test -race ./tests/system/

# Run with memory profiling
go test -memprofile=mem.prof ./tests/system/
```

## Release Criteria

### Pre-Release Checklist

- [ ] All unit tests pass
- [ ] All system tests pass
- [ ] All security tests pass
- [ ] Performance benchmarks met
- [ ] Documentation tests pass
- [ ] User acceptance testing completed
- [ ] Cross-platform compatibility verified
- [ ] Package distribution tested
- [ ] Security audit completed

### Sign-off Requirements

1. **Development Team**: Code review and unit tests
2. **QA Team**: System and integration tests
3. **Security Team**: Security audit and vulnerability assessment
4. **Product Team**: User acceptance testing
5. **DevOps Team**: Distribution and deployment verification

### Release Notes

Include test results in release notes:

```markdown
## Testing Summary

- **Test Coverage**: 85.2%
- **Performance**: All benchmarks passed
- **Security**: No high-severity issues found
- **Platforms Tested**: Linux, macOS, Windows
- **Package Formats**: Docker, Homebrew, Scoop, Snap, DEB, RPM
```

This comprehensive testing approach ensures that gh-notif meets all quality, security, and performance requirements before release.
