# Deployment Guide

This guide covers how to deploy and distribute gh-notif across different platforms and package managers.

## Release Process

### 1. Preparation

Before creating a release:

```bash
# Ensure you're on the main branch
git checkout main
git pull origin main

# Run all tests and checks
make check

# Update version in relevant files
# - CHANGELOG.md
# - packaging files
# - documentation

# Test the release process locally
make release-dry-run
```

### 2. Creating a Release

```bash
# Create and push a version tag
git tag v1.2.3
git push origin v1.2.3

# The GitHub Actions workflow will automatically:
# - Build binaries for all platforms
# - Create release artifacts
# - Update package managers
# - Publish Docker images
# - Generate changelog
```

### 3. Post-Release

After the automated release:

1. Verify all artifacts are created
2. Test installation from package managers
3. Update documentation if needed
4. Announce the release

## Platform-Specific Deployment

### Linux

#### Debian/Ubuntu (APT)

```bash
# The release process automatically creates .deb packages
# Users can install with:
curl -fsSL https://github.com/user/gh-notif/releases/latest/download/gh-notif_amd64.deb -o gh-notif.deb
sudo dpkg -i gh-notif.deb
```

#### RHEL/Fedora (YUM/DNF)

```bash
# The release process automatically creates .rpm packages
# Users can install with:
curl -fsSL https://github.com/user/gh-notif/releases/latest/download/gh-notif-1.0.0-1.x86_64.rpm -o gh-notif.rpm
sudo rpm -i gh-notif.rpm
```

#### Snap

```bash
# Build and publish to Snap Store
snapcraft login
snapcraft upload gh-notif_1.0.0_amd64.snap --release=stable

# Users can install with:
sudo snap install gh-notif
```

#### Flatpak

```bash
# Build and publish to Flathub
# This requires submitting to the Flathub repository
# Users can install with:
flatpak install flathub com.github.user.gh-notif
```

### macOS

#### Homebrew

```bash
# The release process automatically updates the Homebrew formula
# Users can install with:
brew install user/tap/gh-notif

# Or from the main tap (if accepted):
brew install gh-notif
```

#### Manual Installation

```bash
# Download and install manually
curl -L https://github.com/user/gh-notif/releases/latest/download/gh-notif_Darwin_x86_64.tar.gz | tar xz
sudo mv gh-notif /usr/local/bin/
```

### Windows

#### Scoop

```bash
# The release process automatically updates the Scoop manifest
# Users can install with:
scoop bucket add user https://github.com/user/scoop-bucket
scoop install gh-notif
```

#### Chocolatey (Future)

```bash
# Planned support for Chocolatey
choco install gh-notif
```

#### Manual Installation

```powershell
# Download and install manually
Invoke-WebRequest -Uri "https://github.com/user/gh-notif/releases/latest/download/gh-notif_Windows_x86_64.zip" -OutFile "gh-notif.zip"
Expand-Archive -Path "gh-notif.zip" -DestinationPath "C:\Program Files\gh-notif"
# Add to PATH manually
```

## Docker Deployment

### Docker Hub

```bash
# Pull and run the latest image
docker pull ghcr.io/user/gh-notif:latest
docker run --rm -it ghcr.io/user/gh-notif:latest --help

# Run with persistent configuration
docker run --rm -it -v ~/.gh-notif:/root/.gh-notif ghcr.io/user/gh-notif:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gh-notif
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gh-notif
  template:
    metadata:
      labels:
        app: gh-notif
    spec:
      containers:
      - name: gh-notif
        image: ghcr.io/user/gh-notif:latest
        command: ["gh-notif", "watch"]
        env:
        - name: GH_NOTIF_AUTH_TOKEN
          valueFrom:
            secretKeyRef:
              name: gh-notif-secret
              key: token
        volumeMounts:
        - name: config
          mountPath: /root/.gh-notif
      volumes:
      - name: config
        configMap:
          name: gh-notif-config
```

## CI/CD Integration

### GitHub Actions

```yaml
# Example workflow for using gh-notif in CI
name: Notification Check
on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours

jobs:
  check-notifications:
    runs-on: ubuntu-latest
    steps:
    - name: Install gh-notif
      run: |
        curl -fsSL https://raw.githubusercontent.com/user/gh-notif/main/scripts/install.sh | bash
        
    - name: Check notifications
      run: |
        gh-notif auth login --token ${{ secrets.GITHUB_TOKEN }}
        gh-notif list --filter "is:unread" --format json > notifications.json
        
    - name: Process notifications
      run: |
        # Custom processing logic here
        cat notifications.json
```

### Jenkins

```groovy
pipeline {
    agent any
    
    stages {
        stage('Install gh-notif') {
            steps {
                sh 'curl -fsSL https://raw.githubusercontent.com/user/gh-notif/main/scripts/install.sh | bash'
            }
        }
        
        stage('Check Notifications') {
            steps {
                withCredentials([string(credentialsId: 'github-token', variable: 'GITHUB_TOKEN')]) {
                    sh '''
                        gh-notif auth login --token $GITHUB_TOKEN
                        gh-notif list --filter "is:unread" --format json
                    '''
                }
            }
        }
    }
}
```

## Monitoring and Observability

### Health Checks

```bash
# Basic health check
gh-notif --version

# Authentication check
gh-notif auth status

# API connectivity check
gh-notif list --limit 1
```

### Metrics Collection

```bash
# Enable profiling
gh-notif profile --http --port 6060 &

# Collect metrics
curl http://localhost:6060/debug/pprof/heap > heap.prof
curl http://localhost:6060/debug/pprof/profile > cpu.prof
```

### Logging

```bash
# Enable debug logging
export GH_NOTIF_DEBUG=true
gh-notif list --debug

# Log to file
gh-notif list 2>&1 | tee gh-notif.log
```

## Security Considerations

### Token Management

```bash
# Use environment variables for tokens
export GH_NOTIF_AUTH_TOKEN="your-token-here"

# Or use secure credential storage
gh-notif auth login  # Uses system keyring
```

### Network Security

```bash
# Configure proxy if needed
export HTTPS_PROXY="https://proxy.company.com:8080"
export HTTP_PROXY="http://proxy.company.com:8080"

# Configure custom CA certificates
export SSL_CERT_FILE="/path/to/ca-certificates.crt"
```

### Container Security

```dockerfile
# Use non-root user in containers
FROM ghcr.io/user/gh-notif:latest
USER 1000:1000
```

## Troubleshooting

### Common Issues

1. **Installation Failures**
   ```bash
   # Check system requirements
   go version  # Should be 1.20+
   
   # Check network connectivity
   curl -I https://api.github.com
   ```

2. **Authentication Issues**
   ```bash
   # Clear stored credentials
   gh-notif auth logout
   
   # Re-authenticate
   gh-notif auth login
   ```

3. **Performance Issues**
   ```bash
   # Clear cache
   rm -rf ~/.gh-notif-cache
   
   # Enable profiling
   gh-notif profile --cpu
   ```

### Debug Commands

```bash
# Verbose output
gh-notif --debug list

# Check configuration
gh-notif config list

# Test API connectivity
gh-notif list --limit 1 --debug
```

## Rollback Procedures

### Package Managers

```bash
# Homebrew
brew uninstall gh-notif
brew install gh-notif@1.0.0  # Install specific version

# Snap
sudo snap revert gh-notif

# Docker
docker pull ghcr.io/user/gh-notif:v1.0.0  # Use specific tag
```

### Manual Rollback

```bash
# Download previous version
curl -L https://github.com/user/gh-notif/releases/download/v1.0.0/gh-notif_Linux_x86_64.tar.gz | tar xz
sudo mv gh-notif /usr/local/bin/
```

## Support and Maintenance

### Update Notifications

Users can check for updates:

```bash
gh-notif version --check-update
```

### Automated Updates

```bash
# Enable auto-updates (when implemented)
gh-notif config set auto_update true

# Manual update
gh-notif version --update
```
