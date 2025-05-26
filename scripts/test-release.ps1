# Local Release Testing Script
# This script tests the release process locally before pushing to GitHub

Write-Host "Testing gh-notif release process locally..." -ForegroundColor Green

# Check if GoReleaser is installed
if (!(Get-Command goreleaser -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] GoReleaser not found. Please install it first:" -ForegroundColor Red
    Write-Host "   scoop install goreleaser" -ForegroundColor Yellow
    Write-Host "   Or download from: https://github.com/goreleaser/goreleaser/releases" -ForegroundColor Yellow
    exit 1
}

Write-Host "[OK] GoReleaser found" -ForegroundColor Green

# Clean previous builds
Write-Host "[INFO] Cleaning previous builds..." -ForegroundColor Blue
if (Test-Path "dist") {
    Remove-Item -Recurse -Force "dist"
}

# Test GoReleaser configuration
Write-Host "[INFO] Checking GoReleaser configuration..." -ForegroundColor Blue
goreleaser check
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] GoReleaser configuration check failed!" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] GoReleaser configuration is valid" -ForegroundColor Green

# Test build process
Write-Host "[INFO] Testing build process..." -ForegroundColor Blue
goreleaser build --snapshot --clean
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Build successful" -ForegroundColor Green

# List generated files
Write-Host "[INFO] Generated files:" -ForegroundColor Blue
if (Test-Path "dist") {
    Get-ChildItem -Recurse "dist" | Select-Object Name, Length | Format-Table
}

# Test Docker build (if Docker is available)
if (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "[INFO] Testing Docker build..." -ForegroundColor Blue

    # Copy a binary to test Docker build
    $linuxBinary = Get-ChildItem "dist" -Recurse -Filter "*linux_amd64*" | Where-Object { $_.Name -eq "gh-notif" } | Select-Object -First 1
    if ($linuxBinary) {
        Copy-Item $linuxBinary.FullName "gh-notif"
        docker build -t gh-notif-test .
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] Docker build successful" -ForegroundColor Green

            # Test running the Docker image
            Write-Host "[INFO] Testing Docker image..." -ForegroundColor Blue
            docker run --rm gh-notif-test version

            # Clean up
            Remove-Item "gh-notif" -ErrorAction SilentlyContinue
        } else {
            Write-Host "[ERROR] Docker build failed!" -ForegroundColor Red
        }
    } else {
        Write-Host "[WARN] No Linux binary found for Docker test" -ForegroundColor Yellow
    }
} else {
    Write-Host "[WARN] Docker not found, skipping Docker build test" -ForegroundColor Yellow
}

# Test full release process (without publishing)
Write-Host "[INFO] Testing full release process (dry run)..." -ForegroundColor Blue
goreleaser release --snapshot --clean
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Release process failed!" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Release process successful" -ForegroundColor Green

# Show summary
Write-Host "`n[SUMMARY] Test Summary:" -ForegroundColor Cyan
Write-Host "[OK] GoReleaser configuration valid" -ForegroundColor Green
Write-Host "[OK] Build process working" -ForegroundColor Green
Write-Host "[OK] Release process working" -ForegroundColor Green

if (Test-Path "dist") {
    $fileCount = (Get-ChildItem -Recurse "dist").Count
    Write-Host "[INFO] Generated $fileCount files in dist/" -ForegroundColor Blue
}

Write-Host "`n[SUCCESS] All tests passed! Ready to push to GitHub." -ForegroundColor Green
Write-Host "[INFO] To create a release, run:" -ForegroundColor Yellow
Write-Host "   git tag v1.0.6" -ForegroundColor White
Write-Host "   git push origin v1.0.6" -ForegroundColor White
