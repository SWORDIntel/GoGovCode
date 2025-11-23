# CodeGov - Go Version (Linux Compatible)

This is a cross-platform Go implementation of the CodeGov tool for generating and managing code.gov inventory JSON files. This version works on Linux, macOS, Windows, and other Unix-like systems.

## Features

- **Cross-platform**: Works on Linux, macOS, Windows, and other Unix-like systems
- **GitHub API integration**: Fetch repository metadata from GitHub organizations
- **code.gov compliance**: Generate JSON files matching code.gov schema v2.0
- **Schema validation**: Validate generated JSON files
- **Override support**: Apply customizations to auto-generated inventory
- **OAuth authentication**: Secure authentication with GitHub API
- **Pagination support**: Handle large repository lists with automatic pagination

## Prerequisites

- **Go 1.21 or later** - [Install Go](https://go.dev/dl/)
- **GitHub Personal Access Token** (optional but recommended for higher API rate limits)

## Installation

### From Source

```bash
# Clone or download the repository
cd UnixCodeGov

# Download dependencies
go mod download

# Build the CLI tool
go build -o codegov-cli ./cmd/codegov-cli

# Optional: Install to your PATH
sudo mv codegov-cli /usr/local/bin/
```

### Quick Start on Linux/Debian/Ubuntu

```bash
# Install Go (if not already installed)
curl -L https://go.dev/dl/go1.21.0.linux-amd64.tar.gz | sudo tar -xz -C /usr/local

# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone https://github.com/NSACodeGov/CodeGov.git
cd CodeGov
go build -o codegov-cli ./cmd/codegov-cli
```

## Usage

### Set GitHub OAuth Token

```bash
# Set your GitHub personal access token (optional)
export OAUTH_TOKEN=your_40_character_token_here

# Or use the CLI tool
./codegov-cli set-token --token your_40_character_token_here
```

### Generate code.gov JSON

```bash
./codegov-cli generate \
  --orgs "NSACodeGov,18F" \
  --agency "NSA" \
  --email "contact@nsa.gov" \
  --name "NSA Cybersecurity" \
  --url "https://nsa.gov/contact" \
  --phone "1-800-NSA-CYBER" \
  --output code.json
```

**Flags:**
- `--orgs` (required): Comma-separated list of GitHub organization names
- `--agency` (required): Federal agency name
- `--email` (required): Contact email address
- `--name` (optional): Contact person name
- `--url` (optional): Contact URL
- `--phone` (optional): Contact phone number
- `--output` (default: code.json): Output file path
- `--include-private`: Include private repositories (default: false)
- `--include-forks`: Include fork repositories (default: false)

### Validate code.gov JSON

```bash
./codegov-cli validate --input code.json
```

### Test URL Accessibility

```bash
./codegov-cli test-url --url "https://github.com/NSACodeGov/CodeGov"
```

### Apply Overrides

Create an `overrides.json` file:

```json
{
  "overrides": [
    {
      "project": "my-project",
      "action": "replaceproperty",
      "property": "laborHours",
      "value": 100
    },
    {
      "project": "another-project",
      "action": "removeproject"
    }
  ]
}
```

Then apply:

```bash
./codegov-cli override \
  --original code.json \
  --new code-final.json \
  --overrides overrides.json
```

## Command Reference

### generate
Generate code.gov JSON from GitHub organization repositories.

```bash
./codegov-cli generate --orgs ORG_NAME --agency AGENCY_NAME --email EMAIL
```

### validate
Validate a code.gov JSON file for compliance.

```bash
./codegov-cli validate --input code.json
```

### set-token
Set GitHub OAuth token for API authentication.

```bash
./codegov-cli set-token --token YOUR_TOKEN
```

### get-token
Retrieve the currently set OAuth token.

```bash
./codegov-cli get-token
```

### test-token
Test if an OAuth token is valid.

```bash
./codegov-cli test-token --token YOUR_TOKEN
```

### test-url
Test if a URL is accessible.

```bash
./codegov-cli test-url --url https://example.com
```

### override
Apply overrides to code.gov JSON files.

```bash
./codegov-cli override --original code.json --new code-final.json --overrides overrides.json
```

## Library Usage

You can also use CodeGov as a Go library in your own applications:

```go
package main

import (
	"fmt"
	"log"

	"github.com/NSACodeGov/CodeGov/codegov"
)

func main() {
	// Set OAuth token
	codegov.SetOAuthToken("your_token_here")

	// Generate code.gov JSON
	agencyOptions := map[string]string{
		"name":  "My Agency",
		"url":   "https://agency.gov",
		"phone": "1-800-AGENCY",
	}

	codeGov, err := codegov.NewCodeGovJSON(
		[]string{"org1", "org2"},
		"Agency Name",
		"contact@agency.gov",
		agencyOptions,
		false, // include private repos
		false, // include forks
	)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	// Validate
	isValid, errors, err := codegov.TestCodeGovJSONFile("code.json")
	if !isValid {
		for _, e := range errors {
			fmt.Printf("Error: %s\n", e)
		}
	}
}
```

## Project Structure

```
UnixCodeGov/
├── codegov/                    # Main library package
│   ├── types.go               # Data structures
│   └── codegov.go             # Core functions
├── cmd/codegov-cli/           # CLI application
│   └── main.go                # CLI entry point
├── go.mod                      # Go module definition
└── README_GO.md                # This file
```

## API Functions

### Authentication
- `SetOAuthToken(token string) error` - Set GitHub OAuth token
- `GetOAuthToken() string` - Get OAuth token from environment
- `TestOAuthToken(token ...string) bool` - Validate token format

### GitHub Integration
- `GetGitHubRepositories(organization string) ([]GitHubRepository, error)`
- `GetGitHubRepositoryLanguages(languagesURL string) ([]string, error)`
- `GetGitHubRepositoryLicense(org, url, project, branch string) (*License, error)`
- `GetGitHubRepositoryLicenseURL(url, branch string) string`
- `GetGitHubRepositoryDisclaimerURL(url, branch string) string`
- `GetGitHubRepositoryReleaseURL(releasesURL string) (string, error)`

### Code.gov Generation
- `NewCodeGovJSON(...) (*CodeGovJSON, error)` - Generate JSON object
- `NewCodeGovJSONFile(...) error` - Generate and save to file
- `TestCodeGovJSONFile(path string) (bool, []string, error)` - Validate JSON

### Utilities
- `TestURL(url string) bool` - Test URL accessibility
- `InvokeCodeGovJsonOverride(original, new, overrides string) error` - Apply overrides

## Environment Variables

- `OAUTH_TOKEN` - GitHub personal access token (optional)

## Examples

### Example 1: Generate inventory for NSA CodeGov

```bash
export OAUTH_TOKEN=your_github_token
./codegov-cli generate \
  --orgs "NSACodeGov" \
  --agency "NSA" \
  --email "opensource@nsa.gov" \
  --output nsa-code.json
```

### Example 2: Generate for multiple organizations

```bash
./codegov-cli generate \
  --orgs "NSACodeGov,18F,Digital-Service-Desk" \
  --agency "MULTIPLE" \
  --email "contact@example.gov" \
  --name "Federal Open Source" \
  --output federal-code.json
```

### Example 3: With overrides

```bash
# Generate initial inventory
./codegov-cli generate \
  --orgs "MyOrg" \
  --agency "MyAgency" \
  --email "contact@agency.gov" \
  --output code-generated.json

# Create overrides.json with custom values
# Apply overrides
./codegov-cli override \
  --original code-generated.json \
  --new code-final.json \
  --overrides overrides.json

# Validate final output
./codegov-cli validate --input code-final.json
```

## Troubleshooting

### Build Issues

If you encounter build errors:

```bash
# Clean build cache
go clean -cache

# Download dependencies again
go mod download

# Rebuild
go build -o codegov-cli ./cmd/codegov-cli
```

### API Rate Limits

Without authentication, GitHub API has a 60 request/hour limit. With a personal access token:

```bash
export OAUTH_TOKEN=your_token
./codegov-cli generate --orgs "YourOrg" --agency "Agency" --email "contact@example.gov"
```

### Network Issues

If you're behind a proxy, you can set standard Go proxy environment variables:

```bash
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=https://proxy.example.com:8080
./codegov-cli generate ...
```

## Differences from PowerShell Version

- **Cross-platform**: Works on any OS with Go runtime
- **Performance**: Faster execution due to Go compilation
- **Dependencies**: No .NET Framework required
- **Deployment**: Single binary executable
- **Library usage**: Can be imported as a Go module

## Schema Version

Conforms to [code.gov JSON schema v2.0](https://github.com/GSA/code-gov/blob/master/JSON_schema/20211104/)

## License

CC0 1.0 Universal - Same as original PowerShell version

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go conventions
- Tests pass
- Changes are documented
- Commit messages are clear and descriptive

## Support

For issues, questions, or contributions, please visit:
https://github.com/NSACodeGov/CodeGov

## Original Project

This Go version is based on the original PowerShell implementation:
- Original repository: https://github.com/NSACodeGov/CodeGov
- Original language: PowerShell
- Conversion: Go (Linux-compatible, cross-platform version)
