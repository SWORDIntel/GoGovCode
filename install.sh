#!/bin/bash

# CodeGov Go Installation Script for Linux/Unix Systems
# This script sets up the CodeGov Go tool on your system

set -e  # Exit on error

echo "=========================================="
echo "CodeGov Go Installation Script"
echo "=========================================="
echo ""

# Check for Go installation
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed!"
    echo ""
    echo "Please install Go 1.21 or later from: https://go.dev/dl/"
    echo ""
    echo "Quick install on Linux:"
    echo "  curl -L https://go.dev/dl/go1.21.0.linux-amd64.tar.gz | sudo tar -xz -C /usr/local"
    echo "  export PATH=\$PATH:/usr/local/go/bin"
    echo ""
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✓ Found Go $GO_VERSION"
echo ""

# Download dependencies
echo "Downloading Go dependencies..."
go mod download
echo "✓ Dependencies downloaded"
echo ""

# Build the CLI tool
echo "Building codegov-cli..."
go build -o codegov-cli ./cmd/codegov-cli
echo "✓ Built codegov-cli"
echo ""

# Test the CLI
if [ -f codegov-cli ]; then
    echo "Testing codegov-cli..."
    ./codegov-cli help > /dev/null
    echo "✓ CLI is working"
    echo ""

    # Offer to install to PATH
    echo "Installation complete!"
    echo ""
    echo "You can now use CodeGov in two ways:"
    echo ""
    echo "1. Run from current directory:"
    echo "   ./codegov-cli generate --help"
    echo ""
    echo "2. Install to your PATH (recommended):"
    echo "   sudo cp codegov-cli /usr/local/bin/"
    echo "   # Then use from anywhere:"
    echo "   codegov-cli generate --help"
    echo ""
    echo "Next steps:"
    echo "1. Set your GitHub token (optional but recommended):"
    echo "   export OAUTH_TOKEN=your_github_token"
    echo ""
    echo "2. Generate code.gov inventory:"
    echo "   ./codegov-cli generate --orgs YourOrg --agency 'Agency Name' --email 'contact@agency.gov' --output code.json"
    echo ""
    echo "3. For more examples, see README_GO.md"
else
    echo "Error: Failed to build codegov-cli"
    exit 1
fi
