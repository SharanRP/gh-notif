# Contributing to gh-notif

Thank you for considering contributing to gh-notif! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
  - [Development Environment](#development-environment)
  - [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
  - [Creating Issues](#creating-issues)
  - [Making Changes](#making-changes)
  - [Pull Requests](#pull-requests)
  - [Code Review](#code-review)
- [Coding Standards](#coding-standards)
  - [Go Style Guide](#go-style-guide)
  - [Documentation](#documentation)
  - [Testing](#testing)
- [Performance Considerations](#performance-considerations)
- [User Experience Guidelines](#user-experience-guidelines)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. By participating, you are expected to uphold this code. Please report unacceptable behavior.

## Getting Started

### Development Environment

1. **Prerequisites**:
   - Go 1.18 or higher
   - Git
   - A GitHub account

2. **Setting up your development environment**:
   ```bash
   # Clone the repository
   git clone https://github.com/SharanRP/gh-notif.git
   cd gh-notif

   # Install dependencies
   go mod download

   # Build the application
   go build -o gh-notif
   ```

3. **Running tests**:
   ```bash
   # Run all tests
   go test ./...

   # Run tests with coverage
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out -o coverage.html
   ```

### Project Structure

```
gh-notif/
├── cmd/                  # Command-line interface
│   └── gh-notif/         # Main command and subcommands
├── internal/             # Internal packages
│   ├── auth/             # Authentication
│   ├── cache/            # Caching
│   ├── config/           # Configuration management
│   ├── filter/           # Notification filtering
│   ├── github/           # GitHub API client
│   ├── grouping/         # Notification grouping
│   ├── output/           # Output formatting
│   ├── scoring/          # Notification scoring
│   ├── search/           # Search functionality
│   ├── ui/               # Terminal UI
│   ├── tutorial/         # Interactive tutorial
│   ├── watch/            # Watch mode
│   └── wizard/           # Setup wizard
├── docs/                 # Documentation
│   ├── man/              # Man pages
│   └── images/           # Documentation images
├── main.go               # Application entry point
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── README.md             # Project overview
└── CONTRIBUTING.md       # Contribution guidelines
```

## Development Workflow

### Creating Issues

Before starting work on a new feature or bug fix, please check if there's an existing issue. If not, create a new one with:

- A clear title and description
- Steps to reproduce (for bugs)
- Expected behavior
- Actual behavior
- Screenshots or terminal output (if applicable)
- Environment information

### Making Changes

1. **Create a branch**:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes**:
   - Write code that follows the [Go Style Guide](#go-style-guide)
   - Add tests for your changes
   - Update documentation as needed

3. **Commit your changes**:
   ```bash
   git commit -m "Brief description of your changes"
   ```

   Use conventional commit messages:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for test changes
   - `refactor:` for code refactoring
   - `perf:` for performance improvements
   - `chore:` for build process or tooling changes

### Pull Requests

1. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a pull request**:
   - Use a clear title and description
   - Reference any related issues
   - Include screenshots or terminal output if applicable
   - Complete the pull request template

3. **Update your PR**:
   - Respond to review comments
   - Make requested changes
   - Push additional commits to your branch

### Code Review

All submissions require review. We use GitHub pull requests for this purpose.

## Coding Standards

### Go Style Guide

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Run `golint` and `go vet` to check for issues
- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for exported functions, types, and packages

### Documentation

- Document all exported functions, types, and packages
- Keep documentation up-to-date with code changes
- Use examples in documentation
- Update the README.md when adding new features
- Add command-line help for new commands and flags

### Testing

- Write unit tests for all new code
- Aim for high test coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test edge cases and error conditions

## Performance Considerations

gh-notif is designed to be high-performance. Keep these considerations in mind:

- Use concurrency appropriately
- Minimize memory allocations
- Use efficient data structures
- Implement caching where appropriate
- Profile your code to identify bottlenecks
- Consider the impact of your changes on API rate limits

## User Experience Guidelines

- Provide clear, actionable error messages
- Use consistent color schemes and styling
- Implement progressive disclosure of complex features
- Make common tasks easy and efficient
- Consider accessibility in all UI changes
- Provide sensible defaults
- Add examples for new commands

## Release Process

1. **Version Numbering**:
   - We use [Semantic Versioning](https://semver.org/)
   - Format: MAJOR.MINOR.PATCH
   - Increment MAJOR for incompatible API changes
   - Increment MINOR for new functionality in a backward-compatible manner
   - Increment PATCH for backward-compatible bug fixes

2. **Release Checklist**:
   - All tests pass
   - Documentation is up-to-date
   - CHANGELOG.md is updated
   - Version number is updated
   - Release notes are prepared

Thank you for contributing to gh-notif!
