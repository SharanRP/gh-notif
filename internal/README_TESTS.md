# Testing Documentation for gh-notif

This document provides an overview of the testing strategy and instructions for running tests for the gh-notif CLI tool.

## Test Structure

The test suite is organized into the following categories:

1. **Unit Tests**: Test individual components in isolation
   - Located in the same package as the code they test
   - File naming convention: `*_test.go`

2. **Integration Tests**: Test interactions between components
   - Located in the `internal/integration` package
   - Test CLI commands and their interactions with the underlying packages

3. **Test Utilities**: Common utilities for testing
   - Located in the `internal/testutil` package
   - Provides mock implementations, test helpers, etc.

## Test Coverage

The test suite aims to provide comprehensive coverage of:

- Authentication functionality
  - OAuth device flow
  - Token storage and encryption
  - Token refresh
  - Error handling

- Configuration management
  - Reading/writing configuration values
  - Environment variable overrides
  - Configuration validation
  - Default value handling

- CLI commands
  - Auth commands (login, status, logout, refresh)
  - Config commands (get, set, list, export, import)

## Running Tests

### Running All Tests

To run all tests:

```bash
go test ./...
```

### Running Tests with Coverage

To run tests with coverage reporting:

#### Windows

```powershell
.\scripts\run_tests.ps1
```

#### Linux/macOS

```bash
./scripts/run_tests.sh
```

This will:
1. Run all tests with coverage tracking
2. Generate an HTML coverage report
3. Open the report in your default browser
4. Display a summary of the coverage in the terminal

### Running Specific Tests

To run tests in a specific package:

```bash
go test ./internal/auth
```

To run a specific test:

```bash
go test ./internal/auth -run TestRefreshToken
```

## Writing Tests

When writing new tests, follow these guidelines:

1. **Test Naming**: Use the format `Test<FunctionName>` for unit tests and `Test<Feature>` for integration tests.

2. **Test Organization**: Use table-driven tests where appropriate to test multiple scenarios.

3. **Mocking**: Use the mock implementations in `testutil` or create new ones as needed.

4. **Cleanup**: Always clean up resources in tests, especially when creating temporary files or directories.

5. **Error Handling**: Test both success and error cases.

6. **Assertions**: Use the standard `testing` package for assertions.

## Mock Implementations

The test suite includes several mock implementations:

1. **MockStorage**: A mock implementation of the `Storage` interface for testing token storage.

2. **MockGitHubAPI**: A mock HTTP server for testing GitHub API interactions.

3. **MockDeviceFlowServer**: A mock server for testing the OAuth device flow.

## Test Environment

Tests are designed to be isolated and idempotent. They:

1. Create temporary directories for test files
2. Mock external dependencies
3. Restore original values after tests complete
4. Clean up resources even if tests fail

## Continuous Integration

The test suite is designed to be run in CI environments. It:

1. Does not require user interaction
2. Does not depend on external services
3. Runs quickly for fast feedback
4. Provides clear error messages for failures
