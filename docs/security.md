# Security Documentation

This document outlines the security measures, practices, and considerations for gh-notif.

## Security Overview

gh-notif handles sensitive data including GitHub authentication tokens and user notifications. Security is implemented at multiple layers:

1. **Authentication Security**: Secure OAuth2 flow and token management
2. **Data Protection**: Encrypted storage and secure transmission
3. **Input Validation**: Comprehensive input sanitization
4. **Supply Chain Security**: Dependency scanning and verification
5. **Runtime Security**: Secure defaults and minimal privileges

## Authentication Security

### OAuth2 Device Flow

gh-notif uses GitHub's OAuth2 device flow for secure authentication:

```go
// Secure device flow implementation
func (a *Authenticator) DeviceFlow() error {
    // 1. Request device code
    deviceCode, err := a.requestDeviceCode()
    if err != nil {
        return err
    }
    
    // 2. Display user code and open browser
    fmt.Printf("Please visit: %s\n", deviceCode.VerificationURI)
    fmt.Printf("Enter code: %s\n", deviceCode.UserCode)
    
    // 3. Poll for token with exponential backoff
    token, err := a.pollForToken(deviceCode)
    if err != nil {
        return err
    }
    
    // 4. Securely store token
    return a.storeToken(token)
}
```

### Token Storage

Tokens are stored securely using platform-specific credential managers:

#### Windows
- Uses Windows Credential Manager
- Tokens encrypted with user's Windows credentials
- Accessible only to the current user

#### macOS
- Uses macOS Keychain
- Tokens encrypted with user's keychain password
- Integrated with system security

#### Linux
- Uses Secret Service API (GNOME Keyring, KDE Wallet)
- Fallback to encrypted file storage
- File permissions restricted to user only

```go
// Platform-specific secure storage
type SecureStorage interface {
    Store(key, value string) error
    Retrieve(key string) (string, error)
    Delete(key string) error
}

// Implementation selection based on platform
func NewSecureStorage() SecureStorage {
    switch runtime.GOOS {
    case "windows":
        return &WindowsCredentialManager{}
    case "darwin":
        return &MacOSKeychain{}
    case "linux":
        return &LinuxSecretService{}
    default:
        return &EncryptedFileStorage{}
    }
}
```

### Token Validation

All tokens are validated before use:

```go
func (c *Client) validateToken(token string) error {
    // Check token format
    if !isValidTokenFormat(token) {
        return ErrInvalidTokenFormat
    }
    
    // Verify token with GitHub API
    user, err := c.getCurrentUser(token)
    if err != nil {
        return fmt.Errorf("token validation failed: %w", err)
    }
    
    // Check required scopes
    if !hasRequiredScopes(user.Scopes) {
        return ErrInsufficientScopes
    }
    
    return nil
}
```

## Data Protection

### Encryption at Rest

Sensitive data is encrypted when stored locally:

```go
// AES-256-GCM encryption for local data
func encryptData(data []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}
```

### Encryption in Transit

All API communications use TLS 1.2+:

```go
// Secure HTTP client configuration
func newSecureHTTPClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                MinVersion:         tls.VersionTLS12,
                InsecureSkipVerify: false,
                CipherSuites: []uint16{
                    tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
                    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
                    tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
                },
            },
        },
        Timeout: 30 * time.Second,
    }
}
```

### Certificate Pinning

For enhanced security, certificate pinning is implemented:

```go
// Certificate pinning for GitHub API
var githubCertFingerprints = []string{
    "sha256:fingerprint1",
    "sha256:fingerprint2",
}

func verifyGitHubCertificate(cert *x509.Certificate) error {
    fingerprint := sha256.Sum256(cert.Raw)
    fingerprintStr := "sha256:" + hex.EncodeToString(fingerprint[:])
    
    for _, expected := range githubCertFingerprints {
        if fingerprintStr == expected {
            return nil
        }
    }
    
    return ErrCertificatePinningFailed
}
```

## Input Validation

### Filter Validation

All user inputs are validated and sanitized:

```go
// Comprehensive input validation
func validateFilter(filter string) error {
    // Check length limits
    if len(filter) > maxFilterLength {
        return ErrFilterTooLong
    }
    
    // Validate syntax
    if !isValidFilterSyntax(filter) {
        return ErrInvalidFilterSyntax
    }
    
    // Check for injection attempts
    if containsSQLInjection(filter) {
        return ErrSQLInjectionAttempt
    }
    
    if containsXSS(filter) {
        return ErrXSSAttempt
    }
    
    return nil
}

// SQL injection detection
func containsSQLInjection(input string) bool {
    sqlPatterns := []string{
        `(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
        `(?i)(script|javascript|vbscript|onload|onerror)`,
        `['"]\s*;\s*--`,
        `['"]\s*or\s+['"]\w+['"]\s*=\s*['"]\w+['"]`,
    }
    
    for _, pattern := range sqlPatterns {
        if matched, _ := regexp.MatchString(pattern, input); matched {
            return true
        }
    }
    
    return false
}
```

### Command Injection Prevention

All external commands are properly escaped:

```go
// Safe command execution
func executeCommand(command string, args ...string) error {
    // Validate command is in allowlist
    if !isAllowedCommand(command) {
        return ErrCommandNotAllowed
    }
    
    // Escape all arguments
    escapedArgs := make([]string, len(args))
    for i, arg := range args {
        escapedArgs[i] = shellescape.Quote(arg)
    }
    
    cmd := exec.Command(command, escapedArgs...)
    return cmd.Run()
}
```

## Supply Chain Security

### Dependency Management

All dependencies are regularly scanned for vulnerabilities:

```bash
# Vulnerability scanning
govulncheck ./...

# Dependency auditing
go list -m all | nancy sleuth

# License compliance
go-licenses check ./...
```

### Build Security

Builds are secured with:

1. **Reproducible Builds**: Deterministic build process
2. **Signed Releases**: All releases are signed with GPG
3. **SBOM Generation**: Software Bill of Materials for transparency
4. **Container Scanning**: Docker images scanned for vulnerabilities

```yaml
# Secure build pipeline
- name: Generate SBOM
  run: |
    syft packages . -o spdx-json > sbom.json
    
- name: Sign release
  run: |
    gpg --detach-sign --armor gh-notif
    
- name: Scan container
  run: |
    trivy image ghcr.io/user/gh-notif:latest
```

## Runtime Security

### Secure Defaults

The application uses secure defaults:

```go
// Secure configuration defaults
var defaultConfig = Config{
    API: APIConfig{
        Timeout:    30 * time.Second,
        RetryCount: 3,
        UserAgent:  "gh-notif/1.0.0",
    },
    Auth: AuthConfig{
        TokenStorage: "auto", // Use most secure available
        Scopes:       []string{"notifications", "repo:status"},
    },
    Security: SecurityConfig{
        ValidateSSL:     true,
        MinTLSVersion:   "1.2",
        AllowInsecure:   false,
        CertPinning:     true,
    },
}
```

### Privilege Minimization

The application runs with minimal privileges:

- No root/administrator privileges required
- Minimal file system access
- Network access only to GitHub API
- No unnecessary system calls

### Memory Safety

Memory safety measures:

```go
// Secure memory handling
func secureZeroMemory(data []byte) {
    for i := range data {
        data[i] = 0
    }
    runtime.GC()
}

// Secure string handling
type SecureString struct {
    data []byte
}

func (s *SecureString) String() string {
    return string(s.data)
}

func (s *SecureString) Clear() {
    secureZeroMemory(s.data)
}
```

## Security Monitoring

### Audit Logging

Security-relevant events are logged:

```go
// Security audit logging
func auditLog(event string, details map[string]interface{}) {
    logEntry := map[string]interface{}{
        "timestamp": time.Now().UTC(),
        "event":     event,
        "details":   details,
        "user":      getCurrentUser(),
        "ip":        getClientIP(),
    }
    
    logger.Info("security_audit", logEntry)
}

// Usage examples
auditLog("auth_success", map[string]interface{}{
    "method": "oauth2_device_flow",
})

auditLog("auth_failure", map[string]interface{}{
    "method": "token_validation",
    "error":  "invalid_token",
})
```

### Anomaly Detection

Basic anomaly detection for security events:

```go
// Rate limiting for API calls
type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    now := time.Now()
    requests := rl.requests[key]
    
    // Remove old requests
    var recent []time.Time
    for _, req := range requests {
        if now.Sub(req) < time.Hour {
            recent = append(recent, req)
        }
    }
    
    // Check rate limit
    if len(recent) >= maxRequestsPerHour {
        auditLog("rate_limit_exceeded", map[string]interface{}{
            "key": key,
            "requests": len(recent),
        })
        return false
    }
    
    recent = append(recent, now)
    rl.requests[key] = recent
    return true
}
```

## Incident Response

### Security Incident Handling

1. **Detection**: Automated monitoring and user reports
2. **Assessment**: Severity classification and impact analysis
3. **Containment**: Immediate measures to limit damage
4. **Eradication**: Remove the threat and vulnerabilities
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Post-incident review and improvements

### Vulnerability Disclosure

Security vulnerabilities should be reported to:
- Email: security@gh-notif.com
- GitHub Security Advisories
- Coordinated disclosure timeline: 90 days

### Emergency Procedures

In case of security incidents:

1. **Immediate Actions**:
   - Revoke compromised tokens
   - Disable affected features
   - Notify users if necessary

2. **Communication**:
   - Internal team notification
   - User communication plan
   - Public disclosure timeline

3. **Recovery**:
   - Deploy security patches
   - Monitor for continued threats
   - Update security measures

## Security Best Practices

### For Users

1. **Token Management**:
   - Use tokens with minimal required scopes
   - Regularly rotate tokens
   - Never share tokens

2. **System Security**:
   - Keep gh-notif updated
   - Use secure operating systems
   - Enable system firewalls

3. **Network Security**:
   - Use trusted networks
   - Avoid public Wi-Fi for sensitive operations
   - Consider VPN usage

### For Developers

1. **Code Security**:
   - Follow secure coding practices
   - Regular security reviews
   - Use static analysis tools

2. **Dependency Management**:
   - Regular dependency updates
   - Vulnerability scanning
   - License compliance

3. **Testing**:
   - Security test cases
   - Penetration testing
   - Code coverage for security functions
