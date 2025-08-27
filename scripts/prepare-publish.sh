#!/bin/bash

# AI Provider Wrapper - Prepare for Publishing Script
# This script helps prepare the package for publishing to Go package registry

set -e

echo "🚀 Preparing AI Provider Wrapper for publishing..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.19 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
print_status "Go version: $GO_VERSION"

# Check if git is installed
if ! command -v git &> /dev/null; then
    print_error "Git is not installed. Please install Git."
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_warning "Not in a git repository. Initializing..."
    git init
    print_success "Git repository initialized"
fi

# Clean up any build artifacts
print_status "Cleaning up build artifacts..."
go clean -cache
go clean -modcache
rm -f *.test *.out

# Verify go.mod exists and is valid
if [ ! -f "go.mod" ]; then
    print_error "go.mod file not found. Please run 'go mod init' first."
    exit 1
fi

print_status "Validating go.mod..."
go mod verify
print_success "go.mod is valid"

# Tidy up dependencies
print_status "Tidying up dependencies..."
go mod tidy
print_success "Dependencies tidied"

# Run tests
print_status "Running tests..."
if go test ./...; then
    print_success "All tests passed"
else
    print_error "Tests failed. Please fix failing tests before publishing."
    exit 1
fi

# Run go vet
print_status "Running go vet..."
if go vet ./...; then
    print_success "go vet passed"
else
    print_error "go vet found issues. Please fix them before publishing."
    exit 1
fi

# Check for main.go (should not exist in library)
if [ -f "main.go" ]; then
    print_warning "main.go found. Library packages should not have main functions."
    print_status "Consider removing main.go or moving it to examples/"
fi

# Check for required files
print_status "Checking for required files..."

required_files=("README.md" "LICENSE" "go.mod")
for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        print_success "✓ $file exists"
    else
        print_warning "✗ $file missing (recommended for publishing)"
    fi
done

# Check if README has installation instructions
if grep -q "go get" README.md 2>/dev/null; then
    print_success "✓ README contains installation instructions"
else
    print_warning "✗ README should include 'go get' installation instructions"
fi

# Build the package
print_status "Building package..."
if go build ./...; then
    print_success "Package builds successfully"
else
    print_error "Package build failed"
    exit 1
fi

# Check git status
print_status "Checking git status..."
if [ -n "$(git status --porcelain)" ]; then
    print_warning "You have uncommitted changes:"
    git status --short
    echo ""
    print_status "Consider committing changes before publishing"
else
    print_success "Working directory is clean"
fi

# Check if origin remote exists
if git remote get-url origin > /dev/null 2>&1; then
    ORIGIN_URL=$(git remote get-url origin)
    print_success "Git origin: $ORIGIN_URL"
else
    print_warning "No git origin remote found. You'll need to set this up for publishing."
fi

# Summary
echo ""
echo "📋 Pre-publishing Summary:"
echo "=========================="
print_status "Package appears ready for publishing!"
echo ""
echo "Next steps:"
echo "1. Commit any remaining changes: git add . && git commit -m 'Prepare for v1.0.0'"
echo "2. Create and push a version tag: git tag v1.0.0 && git push origin v1.0.0"
echo "3. Create a GitHub release (optional but recommended)"
echo "4. The package will be automatically indexed by pkg.go.dev"
echo ""
echo "To install your published package, users will run:"
echo "go get $(go list -m)"
echo ""
print_success "Preparation complete! 🎉"