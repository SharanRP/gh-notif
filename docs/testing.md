# Testing Strategy

This document outlines the comprehensive testing strategy for gh-notif, covering unit tests, integration tests, performance tests, and quality assurance.

## Testing Philosophy

Our testing approach follows these principles:

1. **Test Pyramid**: More unit tests, fewer integration tests, minimal E2E tests
2. **Fast Feedback**: Tests should run quickly and provide immediate feedback
3. **Reliability**: Tests should be deterministic and not flaky
4. **Coverage**: Aim for >80% code coverage with meaningful tests
5. **Documentation**: Tests serve as living documentation

## Test Categories

### Unit Tests

Unit tests focus on individual functions and methods in isolation.

#### Location
- `*_test.go` files alongside source code
- Test packages mirror source package structure

#### Coverage Areas
- Authentication logic
- Configuration management
- Filtering algorithms
- Notification processing
- API client methods
- Utility functions

#### Example
```go
func TestFilterNotifications(t *testing.T) {
    tests := []struct {
        name         string
        notifications []Notification
        filter       string
        expected     []Notification
    }{
        {
            name: "filter by repository",
            notifications: []Notification{
                {Repository: "owner/repo1"},
                {Repository: "owner/repo2"},
            },
            filter: "repo:owner/repo1",
            expected: []Notification{
                {Repository: "owner/repo1"},
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FilterNotifications(tt.notifications, tt.filter)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### Running Unit Tests
```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/filter

# Run with race detection
go test -race ./...
```

### Integration Tests

Integration tests verify that components work together correctly.

#### Location
- `tests/integration/` directory
- Separate from unit tests

#### Coverage Areas
- GitHub API integration
- Database operations
- Configuration loading
- Command-line interface
- Authentication flow

#### Example
```go
func TestGitHubAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    client := github.NewClient(testToken)
    notifications, err := client.ListNotifications(context.Background())
    
    assert.NoError(t, err)
    assert.NotNil(t, notifications)
}
```

#### Running Integration Tests
```bash
# Run integration tests
go test ./tests/integration/...

# Skip integration tests
go test -short ./...

# Run with environment setup
GITHUB_TOKEN=your_token go test ./tests/integration/...
```

### End-to-End Tests

E2E tests verify complete user workflows.

#### Location
- `tests/e2e/` directory
- Uses real GitHub API with test data

#### Coverage Areas
- Complete authentication flow
- Notification listing and filtering
- Configuration management
- Command-line workflows

#### Example
```go
func TestCompleteWorkflow(t *testing.T) {
    // Setup test environment
    tmpDir := t.TempDir()
    configFile := filepath.Join(tmpDir, "config.yaml")
    
    // Test authentication
    cmd := exec.Command("gh-notif", "auth", "login", "--config", configFile)
    // ... test implementation
    
    // Test listing notifications
    cmd = exec.Command("gh-notif", "list", "--config", configFile)
    // ... test implementation
}
```

### Performance Tests

Performance tests ensure the application meets performance requirements.

#### Benchmarks

```go
func BenchmarkFilterNotifications(b *testing.B) {
    notifications := generateTestNotifications(1000)
    filter := "repo:owner/repo AND is:unread"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        FilterNotifications(notifications, filter)
    }
}

func BenchmarkAPICall(b *testing.B) {
    client := github.NewClient(testToken)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.ListNotifications(context.Background())
    }
}
```

#### Load Tests

```bash
# Run benchmarks
go test -bench=. ./...

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
```

#### Performance Criteria

- API calls should complete within 5 seconds
- Filtering 1000 notifications should take <100ms
- Memory usage should not exceed 100MB for normal operations
- Startup time should be <1 second

### Security Tests

Security tests verify the application handles security concerns properly.

#### Areas Covered
- Input validation
- Authentication security
- Token storage security
- API security

#### Example
```go
func TestInputValidation(t *testing.T) {
    tests := []struct {
        name  string
        input string
        valid bool
    }{
        {"valid filter", "repo:owner/repo", true},
        {"sql injection", "'; DROP TABLE users; --", false},
        {"xss attempt", "<script>alert('xss')</script>", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateFilter(tt.input)
            if tt.valid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

## Test Infrastructure

### Test Data Management

#### Mock Data
```go
// Generate consistent test data
func generateTestNotifications(count int) []Notification {
    notifications := make([]Notification, count)
    for i := 0; i < count; i++ {
        notifications[i] = Notification{
            ID:         fmt.Sprintf("notif-%d", i),
            Repository: fmt.Sprintf("owner/repo-%d", i%10),
            Type:       "PullRequest",
            Unread:     i%2 == 0,
        }
    }
    return notifications
}
```

#### Test Fixtures
```go
// Load test data from files
func loadTestFixture(filename string) []Notification {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    if err != nil {
        panic(err)
    }
    
    var notifications []Notification
    json.Unmarshal(data, &notifications)
    return notifications
}
```

### Mocking and Stubbing

#### HTTP Mocking
```go
func TestAPIClient(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode([]Notification{
            {ID: "test-1", Repository: "owner/repo"},
        })
    }))
    defer server.Close()
    
    client := github.NewClient("test-token")
    client.BaseURL = server.URL
    
    notifications, err := client.ListNotifications(context.Background())
    assert.NoError(t, err)
    assert.Len(t, notifications, 1)
}
```

#### Interface Mocking
```go
type MockGitHubClient struct {
    notifications []Notification
    err          error
}

func (m *MockGitHubClient) ListNotifications(ctx context.Context) ([]Notification, error) {
    return m.notifications, m.err
}
```

### Test Environment Setup

#### Docker Test Environment
```dockerfile
# Dockerfile.test
FROM golang:1.21-alpine

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY . .

RUN go mod download
CMD ["go", "test", "./..."]
```

#### GitHub Actions Test Matrix
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
    go-version: ['1.20', '1.21']
```

## Quality Gates

### Coverage Requirements

- Minimum 80% code coverage
- Critical paths must have 95% coverage
- New code must not decrease overall coverage

### Performance Requirements

- All benchmarks must pass
- No performance regressions >10%
- Memory usage must not increase >20%

### Security Requirements

- All security tests must pass
- No high-severity vulnerabilities
- All inputs must be validated

## Continuous Testing

### Pre-commit Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Run tests
go test ./...
if [ $? -ne 0 ]; then
    echo "Tests failed"
    exit 1
fi

# Run linting
golangci-lint run
if [ $? -ne 0 ]; then
    echo "Linting failed"
    exit 1
fi
```

### CI Pipeline

1. **Fast Tests**: Unit tests and linting (< 5 minutes)
2. **Integration Tests**: API integration tests (< 10 minutes)
3. **Security Tests**: Security scanning (< 5 minutes)
4. **Performance Tests**: Benchmarks and load tests (< 15 minutes)
5. **E2E Tests**: Complete workflow tests (< 20 minutes)

### Test Reporting

- Coverage reports uploaded to Codecov
- Test results displayed in GitHub Actions
- Performance metrics tracked over time
- Security scan results in GitHub Security tab

## Test Maintenance

### Regular Tasks

- Update test data monthly
- Review and update performance benchmarks
- Refresh integration test tokens
- Update security test cases

### Test Cleanup

```bash
# Clean test artifacts
go clean -testcache
rm -f *.prof coverage.out

# Reset test database
rm -f test.db
```

### Debugging Failed Tests

```bash
# Run specific test with verbose output
go test -v -run TestSpecificFunction

# Run with race detection
go test -race -run TestSpecificFunction

# Generate test coverage for specific package
go test -coverprofile=cover.out ./internal/filter
go tool cover -html=cover.out
```
