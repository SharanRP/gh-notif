# Comprehensive End-to-End Test Runner for gh-notif (PowerShell)
# This script runs all test suites and generates a comprehensive report

param(
    [switch]$Verbose,
    [string]$Timeout = "30m",
    [string]$ReportDir = "test-reports",
    [switch]$Help
)

# Configuration
$ErrorActionPreference = "Stop"
$GitHubToken = $env:GITHUB_TOKEN
$ParallelTests = 4

# Test suites
$SystemTests = "tests/system"
$E2ETests = "tests/e2e"
$AcceptanceTests = "tests/acceptance"
$SecurityTests = "tests/security"
$DistributionTests = "tests/distribution"

# Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Show-Help {
    Write-Host "Usage: .\run-e2e-tests.ps1 [options]"
    Write-Host "Options:"
    Write-Host "  -Verbose           Enable verbose output"
    Write-Host "  -Timeout TIMEOUT   Set test timeout (default: 30m)"
    Write-Host "  -ReportDir DIR     Set report directory (default: test-reports)"
    Write-Host "  -Help              Show this help message"
    exit 0
}

function Test-Prerequisites {
    Write-Info "Checking prerequisites..."
    
    # Check Go
    try {
        $null = Get-Command go -ErrorAction Stop
    }
    catch {
        Write-Error "Go is required but not installed"
    }
    
    # Check Git
    try {
        $null = Get-Command git -ErrorAction Stop
    }
    catch {
        Write-Error "Git is required but not installed"
    }
    
    # Check if we're in the right directory
    if (-not (Test-Path "go.mod") -or -not (Test-Path "main.go")) {
        Write-Error "Please run this script from the gh-notif root directory"
    }
    
    # Check GitHub token
    if (-not $GitHubToken) {
        Write-Warn "GITHUB_TOKEN not set - some tests will be skipped"
    }
    
    Write-Success "Prerequisites check passed"
}

function Initialize-TestEnvironment {
    Write-Info "Setting up test environment..."
    
    # Create report directory
    if (Test-Path $ReportDir) {
        Remove-Item $ReportDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $ReportDir -Force | Out-Null
    
    # Build the binary for testing
    Write-Info "Building gh-notif for testing..."
    go build -o gh-notif-test.exe ./main.go
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build test binary"
    }
    
    # Set environment variables
    $env:GH_NOTIF_TEST_BINARY = "$(Get-Location)\gh-notif-test.exe"
    $env:GH_NOTIF_TEST_MODE = "true"
    
    Write-Success "Test environment setup complete"
}

function Invoke-UnitTests {
    Write-Info "Running unit tests..."
    
    $reportFile = "$ReportDir\unit-tests.xml"
    $coverageFile = "$ReportDir\coverage.out"
    $logFile = "$ReportDir\unit-tests.log"
    
    if ($Verbose) {
        go test -v -race -coverprofile="$coverageFile" -timeout="$Timeout" ./... | Tee-Object -FilePath $logFile
    }
    else {
        go test -race -coverprofile="$coverageFile" -timeout="$Timeout" ./... > $logFile 2>&1
    }
    
    $exitCode = $LASTEXITCODE
    
    # Generate coverage report
    if (Test-Path $coverageFile) {
        go tool cover -html="$coverageFile" -o "$ReportDir\coverage.html"
        go tool cover -func="$coverageFile" > "$ReportDir\coverage.txt"
        
        # Extract coverage percentage
        $coverage = (go tool cover -func="$coverageFile" | Select-String "total" | ForEach-Object { $_.Line.Split()[-1] })
        Write-Info "Code coverage: $coverage"
    }
    
    if ($exitCode -eq 0) {
        Write-Success "Unit tests passed"
    }
    else {
        Write-Error "Unit tests failed"
    }
    
    return $exitCode
}

function Invoke-SystemTests {
    Write-Info "Running system tests..."
    
    $logFile = "$ReportDir\system-tests.log"
    
    if (Test-Path $SystemTests) {
        if ($Verbose) {
            go test -v -timeout="$Timeout" "./$SystemTests/..." | Tee-Object -FilePath $logFile
        }
        else {
            go test -timeout="$Timeout" "./$SystemTests/..." > $logFile 2>&1
        }
        
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-Success "System tests passed"
        }
        else {
            Write-Warn "System tests failed or skipped"
        }
        
        return $exitCode
    }
    else {
        Write-Warn "System tests directory not found: $SystemTests"
        return 0
    }
}

function Invoke-E2ETests {
    Write-Info "Running end-to-end tests..."
    
    $logFile = "$ReportDir\e2e-tests.log"
    
    if (Test-Path $E2ETests) {
        if ($Verbose) {
            go test -v -timeout="$Timeout" "./$E2ETests/..." | Tee-Object -FilePath $logFile
        }
        else {
            go test -timeout="$Timeout" "./$E2ETests/..." > $logFile 2>&1
        }
        
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-Success "End-to-end tests passed"
        }
        else {
            Write-Warn "End-to-end tests failed or skipped"
        }
        
        return $exitCode
    }
    else {
        Write-Warn "E2E tests directory not found: $E2ETests"
        return 0
    }
}

function Invoke-SecurityTests {
    Write-Info "Running security tests..."
    
    $logFile = "$ReportDir\security-tests.log"
    
    if (Test-Path $SecurityTests) {
        if ($Verbose) {
            go test -v -timeout="$Timeout" "./$SecurityTests/..." | Tee-Object -FilePath $logFile
        }
        else {
            go test -timeout="$Timeout" "./$SecurityTests/..." > $logFile 2>&1
        }
        
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-Success "Security tests passed"
        }
        else {
            Write-Warn "Security tests failed or skipped"
        }
        
        return $exitCode
    }
    else {
        Write-Warn "Security tests directory not found: $SecurityTests"
        return 0
    }
}

function Invoke-DistributionTests {
    Write-Info "Running distribution tests..."
    
    $logFile = "$ReportDir\distribution-tests.log"
    
    if (Test-Path $DistributionTests) {
        if ($Verbose) {
            go test -v -timeout="$Timeout" "./$DistributionTests/..." | Tee-Object -FilePath $logFile
        }
        else {
            go test -timeout="$Timeout" "./$DistributionTests/..." > $logFile 2>&1
        }
        
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-Success "Distribution tests passed"
        }
        else {
            Write-Warn "Distribution tests failed or skipped"
        }
        
        return $exitCode
    }
    else {
        Write-Warn "Distribution tests directory not found: $DistributionTests"
        return 0
    }
}

function Invoke-PerformanceBenchmarks {
    Write-Info "Running performance benchmarks..."
    
    $benchFile = "$ReportDir\benchmarks.txt"
    
    go test -bench=. -benchmem -timeout="$Timeout" ./... > $benchFile 2>&1
    
    $exitCode = $LASTEXITCODE
    
    if ($exitCode -eq 0) {
        Write-Success "Performance benchmarks completed"
    }
    else {
        Write-Warn "Performance benchmarks failed"
    }
    
    return $exitCode
}

function Invoke-StaticAnalysis {
    Write-Info "Running static analysis..."
    
    # golangci-lint
    if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
        golangci-lint run --out-format=json > "$ReportDir\lint-report.json" 2>&1
        golangci-lint run > "$ReportDir\lint-report.txt" 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Linting passed"
        }
        else {
            Write-Warn "Linting found issues"
        }
    }
    else {
        Write-Warn "golangci-lint not found, skipping linting"
    }
    
    # gosec
    if (Get-Command gosec -ErrorAction SilentlyContinue) {
        gosec -fmt json -out "$ReportDir\security-scan.json" ./...
        gosec -fmt text -out "$ReportDir\security-scan.txt" ./...
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Security scan passed"
        }
        else {
            Write-Warn "Security scan found issues"
        }
    }
    else {
        Write-Warn "gosec not found, skipping security scan"
    }
    
    # govulncheck
    if (Get-Command govulncheck -ErrorAction SilentlyContinue) {
        govulncheck ./... > "$ReportDir\vulnerability-check.txt" 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Vulnerability check passed"
        }
        else {
            Write-Warn "Vulnerability check found issues"
        }
    }
    else {
        Write-Warn "govulncheck not found, skipping vulnerability check"
    }
}

function New-TestReport {
    Write-Info "Generating test report..."
    
    $reportFile = "$ReportDir\test-summary.md"
    
    $content = @"
# gh-notif End-to-End Test Report

**Generated:** $(Get-Date)
**Platform:** $($env:OS) $($env:PROCESSOR_ARCHITECTURE)
**Go Version:** $(go version)

## Test Results Summary

"@
    
    # Add test results
    Get-ChildItem "$ReportDir\*.log" | ForEach-Object {
        $testName = $_.BaseName
        $content += "`n### $testName`n"
        
        $logContent = Get-Content $_.FullName -Raw
        if ($logContent -match "PASS") {
            $content += "✅ **PASSED**`n"
        }
        elseif ($logContent -match "FAIL") {
            $content += "❌ **FAILED**`n"
        }
        else {
            $content += "⚠️ **SKIPPED**`n"
        }
        
        $content += "`n"
    }
    
    # Add coverage information
    if (Test-Path "$ReportDir\coverage.txt") {
        $content += "`n## Code Coverage`n"
        $content += "``````n"
        $content += Get-Content "$ReportDir\coverage.txt" -Raw
        $content += "``````n`n"
    }
    
    # Add benchmark results
    if (Test-Path "$ReportDir\benchmarks.txt") {
        $content += "`n## Performance Benchmarks`n"
        $content += "``````n"
        $benchmarks = Get-Content "$ReportDir\benchmarks.txt" | Where-Object { $_ -match "Benchmark" }
        $content += $benchmarks -join "`n"
        $content += "``````n`n"
    }
    
    Set-Content -Path $reportFile -Value $content
    
    Write-Success "Test report generated: $reportFile"
}

function Clear-TestEnvironment {
    Write-Info "Cleaning up..."
    
    # Remove test binary
    if (Test-Path "gh-notif-test.exe") {
        Remove-Item "gh-notif-test.exe" -Force
    }
    
    # Unset environment variables
    Remove-Item env:GH_NOTIF_TEST_BINARY -ErrorAction SilentlyContinue
    Remove-Item env:GH_NOTIF_TEST_MODE -ErrorAction SilentlyContinue
    
    Write-Success "Cleanup complete"
}

# Main execution
function Main {
    if ($Help) {
        Show-Help
    }
    
    Write-Info "Starting comprehensive end-to-end testing for gh-notif"
    
    # Run test phases
    Test-Prerequisites
    Initialize-TestEnvironment
    
    # Track test results
    $unitResult = 0
    $systemResult = 0
    $e2eResult = 0
    $securityResult = 0
    $distributionResult = 0
    
    # Run all test suites
    try { $unitResult = Invoke-UnitTests } catch { $unitResult = 1 }
    try { $systemResult = Invoke-SystemTests } catch { $systemResult = 1 }
    try { $e2eResult = Invoke-E2ETests } catch { $e2eResult = 1 }
    try { $securityResult = Invoke-SecurityTests } catch { $securityResult = 1 }
    try { $distributionResult = Invoke-DistributionTests } catch { $distributionResult = 1 }
    
    # Run additional checks
    Invoke-PerformanceBenchmarks
    Invoke-StaticAnalysis
    
    # Generate report
    New-TestReport
    
    # Cleanup
    Clear-TestEnvironment
    
    # Final summary
    Write-Info "Test execution complete"
    Write-Info "Results:"
    Write-Info "  Unit tests: $(if ($unitResult -eq 0) { "PASS" } else { "FAIL" })"
    Write-Info "  System tests: $(if ($systemResult -eq 0) { "PASS" } else { "FAIL" })"
    Write-Info "  E2E tests: $(if ($e2eResult -eq 0) { "PASS" } else { "FAIL" })"
    Write-Info "  Security tests: $(if ($securityResult -eq 0) { "PASS" } else { "FAIL" })"
    Write-Info "  Distribution tests: $(if ($distributionResult -eq 0) { "PASS" } else { "FAIL" })"
    
    # Exit with error if any critical tests failed
    if ($unitResult -ne 0) {
        Write-Error "Unit tests failed - this is a critical failure"
    }
    
    if ($systemResult -ne 0 -or $e2eResult -ne 0 -or $securityResult -ne 0) {
        Write-Warn "Some test suites failed - review the reports"
        exit 1
    }
    
    Write-Success "All tests completed successfully!"
}

# Run main function
Main
