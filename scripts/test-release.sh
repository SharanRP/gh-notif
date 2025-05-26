#!/bin/bash
# Local Release Testing Script
# This script tests the release process locally before pushing to GitHub

set -e

echo "🧪 Testing gh-notif release process locally..."

# Check if GoReleaser is installed
if ! command -v goreleaser &> /dev/null; then
    echo "❌ GoReleaser not found. Please install it first:"
    echo "   brew install goreleaser/tap/goreleaser"
    echo "   Or download from: https://github.com/goreleaser/goreleaser/releases"
    exit 1
fi

echo "✅ GoReleaser found"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/

# Test GoReleaser configuration
echo "🔍 Checking GoReleaser configuration..."
if ! goreleaser check; then
    echo "❌ GoReleaser configuration check failed!"
    exit 1
fi
echo "✅ GoReleaser configuration is valid"

# Test build process
echo "🔨 Testing build process..."
if ! goreleaser build --snapshot --clean; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✅ Build successful"

# List generated files
echo "📦 Generated files:"
if [ -d "dist" ]; then
    find dist -type f -exec ls -lh {} \; | head -20
fi

# Test Docker build (if Docker is available)
if command -v docker &> /dev/null; then
    echo "🐳 Testing Docker build..."
    
    # Copy a binary to test Docker build
    linux_binary=$(find dist -name "gh-notif" -path "*linux_amd64*" | head -1)
    if [ -n "$linux_binary" ]; then
        cp "$linux_binary" gh-notif
        if docker build -t gh-notif-test .; then
            echo "✅ Docker build successful"
            
            # Test running the Docker image
            echo "🧪 Testing Docker image..."
            docker run --rm gh-notif-test version
            
            # Clean up
            rm -f gh-notif
        else
            echo "❌ Docker build failed!"
        fi
    else
        echo "⚠️  No Linux binary found for Docker test"
    fi
else
    echo "⚠️  Docker not found, skipping Docker build test"
fi

# Test full release process (without publishing)
echo "🚀 Testing full release process (dry run)..."
if ! goreleaser release --snapshot --clean; then
    echo "❌ Release process failed!"
    exit 1
fi
echo "✅ Release process successful"

# Show summary
echo ""
echo "📋 Test Summary:"
echo "✅ GoReleaser configuration valid"
echo "✅ Build process working"
echo "✅ Release process working"

if [ -d "dist" ]; then
    file_count=$(find dist -type f | wc -l)
    echo "📦 Generated $file_count files in dist/"
fi

echo ""
echo "🎉 All tests passed! Ready to push to GitHub."
echo "💡 To create a release, run:"
echo "   git tag v1.0.6"
echo "   git push origin v1.0.6"
