# Test coverage script for PowerShell

# Create coverage directory if it doesn't exist
$coverageDir = "coverage"
if (-not (Test-Path $coverageDir)) {
    New-Item -ItemType Directory -Path $coverageDir | Out-Null
}

# Run tests with coverage
Write-Host "Running tests with coverage..."
go test ./... -coverprofile="$coverageDir/coverage.out" -covermode=atomic

# Generate HTML coverage report
Write-Host "Generating HTML coverage report..."
go tool cover -html="$coverageDir/coverage.out" -o="$coverageDir/coverage.html"

# Calculate coverage percentage
$coverageOutput = go tool cover -func="$coverageDir/coverage.out"
$totalLine = $coverageOutput | Select-String -Pattern "total:"
if ($totalLine -match "total:\s+\(statements\)\s+(\d+\.\d+)%") {
    $coveragePercentage = $matches[1]
    Write-Host "Total coverage: $coveragePercentage%"
} else {
    Write-Host "Could not determine total coverage"
}

# Open the coverage report in the default browser
Write-Host "Opening coverage report in browser..."
Invoke-Item "$coverageDir/coverage.html"
