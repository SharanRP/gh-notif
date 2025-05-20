# PowerShell script to run tests with coverage reporting

# Create a directory for coverage reports if it doesn't exist
$coverageDir = "coverage"
if (-not (Test-Path $coverageDir)) {
    New-Item -ItemType Directory -Path $coverageDir | Out-Null
}

# Clean up any previous coverage files
Remove-Item -Path "$coverageDir\*" -Force -ErrorAction SilentlyContinue

# Run tests with coverage
Write-Host "Running tests with coverage..."
go test ./... -coverprofile="$coverageDir\coverage.out" -covermode=atomic

# Check if the tests passed
if ($LASTEXITCODE -ne 0) {
    Write-Host "Tests failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}

# Generate HTML coverage report
Write-Host "Generating HTML coverage report..."
go tool cover -html="$coverageDir\coverage.out" -o="$coverageDir\coverage.html"

# Generate coverage summary
Write-Host "Generating coverage summary..."
$coverageReport = go tool cover -func="$coverageDir\coverage.out"
$coverageReport | Out-File -FilePath "$coverageDir\coverage_summary.txt"

# Extract the total coverage percentage
$totalCoverage = ($coverageReport | Select-String -Pattern "total:.*statements.*").Matches.Value
Write-Host "Coverage: $totalCoverage" -ForegroundColor Green

# Open the coverage report in the default browser
Write-Host "Opening coverage report in browser..."
Invoke-Item "$coverageDir\coverage.html"

Write-Host "Test run completed successfully!" -ForegroundColor Green
