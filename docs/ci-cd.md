# CI/CD Documentation

This document describes the continuous integration and deployment setup for gh-notif.

## Overview

The gh-notif project uses GitHub Actions for CI/CD with the following components:

- **Continuous Integration**: Automated testing, linting, and security scanning
- **Cross-Platform Builds**: Automated builds for Linux, macOS, and Windows
- **Package Distribution**: Automated package creation for multiple package managers
- **Release Management**: Automated releases with GoReleaser
- **Quality Assurance**: Code coverage, security scanning, and vulnerability checks

## Workflows

### CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request to main/develop branches:

1. **Test Matrix**: Tests on Ubuntu, Windows, and macOS with Go 1.20 and 1.21
2. **Linting**: Uses golangci-lint for code quality checks
3. **Format Check**: Ensures code is properly formatted with gofmt
4. **Security Scan**: Runs gosec for security vulnerability detection
5. **Vulnerability Check**: Uses govulncheck for dependency vulnerabilities
6. **Build**: Cross-platform builds for all supported platforms
7. **Integration Tests**: End-to-end testing on all platforms

### Release Workflow (`.github/workflows/release.yml`)

Triggered on version tags (v*):

1. **GoReleaser**: Creates release artifacts for all platforms
2. **Docker Images**: Builds and publishes Docker images
3. **Package Updates**: Updates Homebrew and Scoop packages
4. **Package Publishing**: Creates and publishes DEB/RPM packages
5. **Notifications**: Sends release notifications

## Build Configuration

### GoReleaser (`.goreleaser.yml`)

Handles:
- Cross-platform binary builds
- Archive creation (tar.gz for Unix, zip for Windows)
- Checksum generation
- Changelog generation
- Docker image builds
- Package manager integrations

### Docker

Two Dockerfiles:
- `Dockerfile`: Production image based on scratch
- `Dockerfile.dev`: Development image with build tools

## Package Distribution

### Homebrew (macOS/Linux)

- Formula: `packaging/homebrew/gh-notif.rb`
- Repository: `user/homebrew-tap`
- Auto-updated via GitHub Actions

### Scoop (Windows)

- Manifest: `packaging/scoop/gh-notif.json`
- Repository: `user/scoop-bucket`
- Auto-updated via GitHub Actions

### Snap (Linux)

- Configuration: `packaging/snap/snapcraft.yaml`
- Published to Snap Store
- Supports amd64 and arm64

### Flatpak (Linux)

- Manifest: `packaging/flatpak/com.github.user.gh-notif.yml`
- Published to Flathub
- Includes desktop integration

### Debian/Ubuntu (APT)

- Created via nfpms in GoReleaser
- Supports amd64 and arm64
- Includes man pages and completions

### RHEL/Fedora (YUM/DNF)

- Created via nfpms in GoReleaser
- Supports amd64 and arm64
- Includes man pages and completions

## Version Management

### Semantic Versioning

The project follows [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- Breaking changes increment MAJOR
- New features increment MINOR
- Bug fixes increment PATCH

### Release Process

1. **Create Release Branch**: `git checkout -b release/v1.2.3`
2. **Update Version**: Update version in relevant files
3. **Update Changelog**: Add release notes to CHANGELOG.md
4. **Create PR**: Submit pull request for review
5. **Merge**: Merge to main branch
6. **Tag Release**: `git tag v1.2.3 && git push origin v1.2.3`
7. **Automated Release**: GitHub Actions handles the rest

### Update Checking

The application includes built-in update checking:
- Checks GitHub releases API
- Compares semantic versions
- Provides update notifications
- Supports self-update (planned)

## Quality Assurance

### Code Coverage

- Measured with `go test -coverprofile`
- Reported to Codecov
- Target: >80% coverage
- Fails CI if coverage drops significantly

### Static Analysis

#### golangci-lint

Configuration in `.golangci.yml`:
- 30+ linters enabled
- Custom rules for the project
- Excludes for test files and generated code

#### gosec

Security-focused static analysis:
- Detects common security issues
- Integrates with GitHub Security tab
- Generates SARIF reports

### Dependency Management

#### govulncheck

- Scans for known vulnerabilities
- Checks both direct and indirect dependencies
- Runs on every CI build

#### Dependabot

- Automated dependency updates
- Security vulnerability alerts
- Automatic PR creation for updates

## Local Development

### Setup

```bash
# Install development tools
make dev-setup

# Run all checks locally
make check

# Run CI pipeline locally
make ci
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Generate completions
make completions

# Generate documentation
make docs
```

### Docker Development

```bash
# Build development image
make docker-build-dev

# Run tests in Docker
make docker-test

# Run interactive development container
docker run -it --rm -v $(pwd):/app gh-notif:dev bash
```

## Secrets and Configuration

### Required Secrets

- `GITHUB_TOKEN`: Automatically provided by GitHub
- `HOMEBREW_TAP_GITHUB_TOKEN`: For updating Homebrew formula
- `SCOOP_BUCKET_GITHUB_TOKEN`: For updating Scoop manifest

### Environment Variables

- `GO_VERSION`: Go version for builds (default: 1.21)
- `REGISTRY`: Container registry (default: ghcr.io)

## Monitoring and Alerts

### Build Status

- GitHub Actions status badges
- Go Report Card integration
- License and version badges

### Release Notifications

- GitHub Releases
- Package manager notifications
- Docker Hub webhooks

## Troubleshooting

### Common Issues

1. **Build Failures**: Check Go version compatibility
2. **Test Failures**: Ensure all dependencies are available
3. **Release Failures**: Verify all secrets are configured
4. **Package Updates**: Check token permissions

### Debug Commands

```bash
# Test GoReleaser locally
make release-dry-run

# Check linting issues
make lint

# Run security scan
make security

# Check vulnerabilities
make vuln
```

## Performance Monitoring

### Benchmarking

The project includes comprehensive benchmarking:

```bash
# Run benchmarks
make bench

# Profile CPU usage
gh-notif profile --cpu --duration 60

# Profile memory usage
gh-notif profile --memory --duration 60

# Run HTTP profiling server
gh-notif profile --http --port 6060
```

### Metrics Collection

- Build time tracking
- Test execution time
- Binary size monitoring
- Memory usage profiling
- API response time tracking

## Security

### Supply Chain Security

- Dependency scanning with govulncheck
- Container image scanning
- SBOM (Software Bill of Materials) generation
- Signed releases with checksums

### Code Security

- Static analysis with gosec
- Secret scanning
- License compliance checking
- Vulnerability disclosure process

## Compliance

### Licensing

- MIT License
- License headers in source files
- Third-party license tracking
- FOSSA integration for compliance

### Documentation

- API documentation
- User guides
- Developer documentation
- Security documentation

## Future Improvements

- [ ] Add more package managers (Chocolatey, AUR)
- [ ] Implement self-update functionality
- [ ] Add performance regression testing
- [ ] Implement canary releases
- [ ] Add integration with package registries
- [ ] Implement automated rollback on failures
- [ ] Add telemetry and usage analytics
- [ ] Implement A/B testing for features
- [ ] Add chaos engineering tests
- [ ] Implement blue-green deployments
