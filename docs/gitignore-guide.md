# .gitignore Guide

This document explains the comprehensive .gitignore configuration for the gh-notif project.

## Overview

The .gitignore file is designed to exclude all generated files, build artifacts, temporary files, and sensitive data from version control. This ensures a clean repository and prevents accidental commits of sensitive information.

## Categories of Ignored Files

### 1. Binaries and Executables

```gitignore
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
gh-notif
gh-notif-*
```

**Purpose**: Excludes compiled binaries and platform-specific executables.

**Files affected**:
- `gh-notif` (Linux/macOS binary)
- `gh-notif.exe` (Windows binary)
- `gh-notif-*` (Variant binaries)
- Platform-specific libraries

### 2. Test Artifacts

```gitignore
# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
```

**Purpose**: Excludes test binaries and coverage output files.

**Files affected**:
- Test binaries created with `go test -c`
- Coverage reports (`coverage.out`)

### 3. Dependencies

```gitignore
# Dependency directories
vendor/
```

**Purpose**: Excludes vendored dependencies (when using Go modules, vendor/ is optional).

### 4. Go Workspace Files

```gitignore
# Go workspace file
go.work
go.work.sum
```

**Purpose**: Excludes Go workspace files that are typically local to development environment.

### 5. IDE and Editor Files

```gitignore
# IDE specific files
.idea/          # IntelliJ IDEA
.vscode/        # Visual Studio Code
*.swp           # Vim swap files
*.swo           # Vim swap files
*~              # Backup files
.vim/           # Vim configuration
.emacs.d/       # Emacs configuration
.sublime-*      # Sublime Text
```

**Purpose**: Excludes IDE and editor configuration files that are specific to individual developers.

### 6. Operating System Files

```gitignore
# OS specific files
.DS_Store       # macOS
.DS_Store?      # macOS
._*             # macOS
.Spotlight-V100 # macOS
.Trashes        # macOS
ehthumbs.db     # Windows
Thumbs.db       # Windows
Desktop.ini     # Windows
```

**Purpose**: Excludes OS-generated files that have no relevance to the project.

### 7. Application-Specific Files

```gitignore
# Application specific files
.gh-notif.yaml      # Main config file
.gh-notif.yml       # Alternative config file
.gh-notif-token.json # Token storage
.gh-notif-token.enc # Encrypted token
.gh-notif-key       # Encryption key
.gh-notif-config/   # Config directory
.gh-notif-cache/    # Cache directory
.gh-notif-history/  # History directory
.gh-notif-badger/   # BadgerDB files
.gh-notif-bolt/     # BoltDB files
.gh-notif-data/     # Data directory
.gh-notif-logs/     # Log directory
```

**Purpose**: Excludes application configuration, cache, and data files that contain user-specific or sensitive information.

### 8. Database Files

```gitignore
# Database files
*.vlog          # BadgerDB value log
*.sst           # BadgerDB SST files
*.mem           # BadgerDB memory files
MANIFEST        # BadgerDB manifest
KEYREGISTRY     # BadgerDB key registry
DISCARD         # BadgerDB discard file
*.db            # Generic database files
*.bolt          # BoltDB files
*.badger        # BadgerDB files
*.sqlite        # SQLite files
*.sqlite3       # SQLite files
```

**Purpose**: Excludes database files created by caching and storage systems.

### 9. Profiling Files

```gitignore
# Profiling files
*.prof          # Generic profile files
*.pprof         # Go pprof files
cpu.prof        # CPU profile
mem.prof        # Memory profile
heap.prof       # Heap profile
goroutine.prof  # Goroutine profile
block.prof      # Block profile
mutex.prof      # Mutex profile
trace.out       # Execution trace
pprof/          # Profile directory
pprof.*         # Profile files
http-profile/   # HTTP profile directory
```

**Purpose**: Excludes performance profiling files generated during development and testing.

### 10. Temporary Files

```gitignore
# Temporary files
*.tmp           # Temporary files
*.temp          # Temporary files
*.log           # Log files
*.bak           # Backup files
*.backup        # Backup files
*.orig          # Original files
main_*.go.bak   # Go backup files
test_*.go.bak   # Test backup files
```

**Purpose**: Excludes temporary and backup files created during development.

### 11. Build Artifacts

```gitignore
# Build artifacts
gh-notif            # Main binary
gh-notif.exe        # Windows binary
gh-notif-ui         # UI variant
gh-notif-ui.exe     # UI variant (Windows)
gh-notif-simple     # Simple variant
gh-notif-simple.exe # Simple variant (Windows)
dist/               # Distribution directory
build/              # Build directory
```

**Purpose**: Excludes compiled binaries and build output directories.

### 12. Coverage Reports

```gitignore
# Coverage reports
coverage/       # Coverage directory
coverage.out    # Go coverage output
coverage.html   # HTML coverage report
coverage.xml    # XML coverage report
coverage.json   # JSON coverage report
*.cover         # Coverage files
```

**Purpose**: Excludes test coverage reports and related files.

### 13. Generated Files

```gitignore
# Generated files
completions/        # Shell completions
docs/man/          # Man pages
*.1                # Man page section 1
*.8                # Man page section 8
*.generated.go     # Generated Go files
*_generated.go     # Generated Go files
*.pb.go            # Protocol buffer files
```

**Purpose**: Excludes files generated by build processes, documentation generation, and code generation tools.

### 14. Release Artifacts

```gitignore
# Release artifacts
*.tar.gz        # Compressed archives
*.zip           # ZIP archives
*.deb           # Debian packages
*.rpm           # RPM packages
*.apk           # Alpine packages
*.dmg           # macOS disk images
*.msi           # Windows installers
checksums.txt   # Checksum file
sbom.json       # Software Bill of Materials
```

**Purpose**: Excludes release packages and distribution artifacts.

### 15. Security Files

```gitignore
# Security files
*.key           # Private keys
*.pem           # PEM files
*.crt           # Certificates
*.p12           # PKCS#12 files
*.pfx           # PFX files
secrets.yaml    # Secrets file
secrets.yml     # Secrets file
.env            # Environment variables
.env.local      # Local environment
.env.*.local    # Environment variants
```

**Purpose**: Excludes sensitive security files and credentials.

## Best Practices

### 1. Regular Review

- Review .gitignore periodically
- Add new patterns as the project evolves
- Remove obsolete patterns

### 2. Testing

Test .gitignore effectiveness:

```bash
# Check what would be committed
git add . --dry-run

# Check ignored files
git status --ignored

# Check if specific file is ignored
git check-ignore filename
```

### 3. Global vs Local

- Use project .gitignore for project-specific files
- Use global .gitignore for personal preferences:

```bash
# Set global gitignore
git config --global core.excludesfile ~/.gitignore_global
```

### 4. Documentation

- Document why specific patterns are included
- Explain application-specific patterns
- Keep this guide updated

## Troubleshooting

### File Already Tracked

If a file is already tracked but should be ignored:

```bash
# Remove from tracking but keep local file
git rm --cached filename

# Remove directory from tracking
git rm -r --cached directory/

# Commit the removal
git commit -m "Remove tracked file that should be ignored"
```

### Pattern Not Working

If a .gitignore pattern isn't working:

1. Check if file is already tracked
2. Verify pattern syntax
3. Test with `git check-ignore`
4. Clear git cache if needed:

```bash
git rm -r --cached .
git add .
git commit -m "Fix .gitignore"
```

### Debugging Patterns

```bash
# Test if file would be ignored
git check-ignore -v filename

# Show all ignored files
git status --ignored --porcelain

# Show gitignore rules affecting a file
git check-ignore -v --no-index filename
```

## Maintenance

### Adding New Patterns

When adding new patterns:

1. Test the pattern first
2. Document the purpose
3. Consider impact on existing files
4. Update this guide

### Removing Patterns

When removing patterns:

1. Ensure files should be tracked
2. Check for sensitive data
3. Update documentation
4. Communicate changes to team

## Security Considerations

### Sensitive Data

Never commit:
- API keys or tokens
- Passwords or credentials
- Private keys or certificates
- Personal configuration files
- Database files with real data

### Recovery

If sensitive data is accidentally committed:

1. Remove from current commit
2. Rewrite history if necessary
3. Rotate compromised credentials
4. Update .gitignore to prevent recurrence

```bash
# Remove sensitive file from history
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch path/to/sensitive/file' \
  --prune-empty --tag-name-filter cat -- --all
```
