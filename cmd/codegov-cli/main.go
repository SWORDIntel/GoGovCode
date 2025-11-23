package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NSACodeGov/CodeGov/codegov"
)

func main() {
	var (
		generateCmd     = flag.NewFlagSet("generate", flag.ExitOnError)
		validateCmd     = flag.NewFlagSet("validate", flag.ExitOnError)
		setTokenCmd     = flag.NewFlagSet("set-token", flag.ExitOnError)
		getTokenCmd     = flag.NewFlagSet("get-token", flag.ExitOnError)
		testTokenCmd    = flag.NewFlagSet("test-token", flag.ExitOnError)
		testURLCmd      = flag.NewFlagSet("test-url", flag.ExitOnError)
		overrideCmd     = flag.NewFlagSet("override", flag.ExitOnError)
	)

	// generate command flags
	generateOrgs := generateCmd.String("orgs", "", "Comma-separated list of GitHub organizations")
	generateAgency := generateCmd.String("agency", "", "Agency name")
	generateEmail := generateCmd.String("email", "", "Contact email")
	generateName := generateCmd.String("name", "", "Contact name (optional)")
	generateURL := generateCmd.String("url", "", "Contact URL (optional)")
	generatePhone := generateCmd.String("phone", "", "Contact phone (optional)")
	generateOutput := generateCmd.String("output", "code.json", "Output file path")
	generatePrivate := generateCmd.Bool("include-private", false, "Include private repositories")
	generateForks := generateCmd.Bool("include-forks", false, "Include fork repositories")

	// validate command flags
	validateInput := validateCmd.String("input", "", "Input JSON file to validate")

	// set-token command flags
	setToken := setTokenCmd.String("token", "", "GitHub OAuth token")

	// test-token command flags
	testToken := testTokenCmd.String("token", "", "GitHub OAuth token to test (uses env var if not provided)")

	// test-url command flags
	testURL := testURLCmd.String("url", "", "URL to test")

	// override command flags
	overrideOriginal := overrideCmd.String("original", "", "Original code.gov JSON file")
	overrideNew := overrideCmd.String("new", "", "New code.gov JSON file")
	overrideFile := overrideCmd.String("overrides", "", "Overrides JSON file")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		if *generateOrgs == "" || *generateAgency == "" || *generateEmail == "" {
			fmt.Println("Error: --orgs, --agency, and --email are required")
			generateCmd.PrintDefaults()
			os.Exit(1)
		}

		agencyOptions := make(map[string]string)
		if *generateName != "" {
			agencyOptions["name"] = *generateName
		}
		if *generateURL != "" {
			agencyOptions["url"] = *generateURL
		}
		if *generatePhone != "" {
			agencyOptions["phone"] = *generatePhone
		}

		orgs := strings.Split(*generateOrgs, ",")
		for i := range orgs {
			orgs[i] = strings.TrimSpace(orgs[i])
		}

		fmt.Printf("Generating code.gov JSON for organizations: %v\n", orgs)
		fmt.Printf("Agency: %s\n", *generateAgency)

		if err := codegov.NewCodeGovJSONFile(orgs, *generateAgency, *generateEmail, agencyOptions, *generatePrivate, *generateForks, *generateOutput); err != nil {
			log.Fatalf("Error generating code.gov JSON: %v\n", err)
		}

		fmt.Printf("Successfully generated code.gov JSON: %s\n", *generateOutput)

	case "validate":
		validateCmd.Parse(os.Args[2:])
		if *validateInput == "" {
			fmt.Println("Error: --input is required")
			validateCmd.PrintDefaults()
			os.Exit(1)
		}

		fmt.Printf("Validating code.gov JSON: %s\n", *validateInput)

		isValid, errors, err := codegov.TestCodeGovJSONFile(*validateInput)
		if err != nil {
			log.Fatalf("Error validating JSON: %v\n", err)
		}

		if isValid {
			fmt.Println("✓ JSON is valid")
		} else {
			fmt.Println("✗ JSON is invalid:")
			for _, e := range errors {
				fmt.Printf("  - %s\n", e)
			}
			os.Exit(1)
		}

	case "set-token":
		setTokenCmd.Parse(os.Args[2:])
		if *setToken == "" {
			fmt.Println("Error: --token is required")
			setTokenCmd.PrintDefaults()
			os.Exit(1)
		}

		if err := codegov.SetOAuthToken(*setToken); err != nil {
			log.Fatalf("Error setting OAuth token: %v\n", err)
		}

		fmt.Println("OAuth token set successfully")

	case "get-token":
		getTokenCmd.Parse(os.Args[2:])
		token := codegov.GetOAuthToken()
		if token == "" {
			fmt.Println("No OAuth token found")
		} else {
			fmt.Printf("OAuth token: %s\n", token)
		}

	case "test-token":
		testTokenCmd.Parse(os.Args[2:])
		var tokenToTest string

		if *testToken != "" {
			tokenToTest = *testToken
		}

		if codegov.TestOAuthToken(tokenToTest) {
			fmt.Println("✓ Token is valid")
		} else {
			fmt.Println("✗ Token is invalid or not set")
			os.Exit(1)
		}

	case "test-url":
		testURLCmd.Parse(os.Args[2:])
		if *testURL == "" {
			fmt.Println("Error: --url is required")
			testURLCmd.PrintDefaults()
			os.Exit(1)
		}

		if codegov.TestURL(*testURL) {
			fmt.Printf("✓ URL is accessible: %s\n", *testURL)
		} else {
			fmt.Printf("✗ URL is not accessible: %s\n", *testURL)
			os.Exit(1)
		}

	case "override":
		overrideCmd.Parse(os.Args[2:])
		if *overrideOriginal == "" || *overrideNew == "" || *overrideFile == "" {
			fmt.Println("Error: --original, --new, and --overrides are required")
			overrideCmd.PrintDefaults()
			os.Exit(1)
		}

		fmt.Printf("Applying overrides from %s\n", *overrideFile)

		if err := codegov.InvokeCodeGovJsonOverride(*overrideOriginal, *overrideNew, *overrideFile); err != nil {
			log.Fatalf("Error applying overrides: %v\n", err)
		}

		fmt.Printf("Successfully applied overrides: %s\n", *overrideNew)

	case "-h", "--help", "help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`CodeGov - Generate and manage code.gov inventory JSON files

Usage:
  codegov-cli [command] [flags]

Commands:
  generate      Generate code.gov JSON from GitHub organizations
  validate      Validate a code.gov JSON file
  set-token     Set GitHub OAuth token
  get-token     Get GitHub OAuth token
  test-token    Test GitHub OAuth token validity
  test-url      Test if a URL is accessible
  override      Apply overrides to code.gov JSON
  help          Show this help message

Examples:
  # Set GitHub OAuth token
  codegov-cli set-token --token YOUR_TOKEN

  # Generate code.gov JSON
  codegov-cli generate \
    --orgs "NSACodeGov,18F" \
    --agency "NSA" \
    --email "contact@nsa.gov" \
    --name "NSA Cybersecurity" \
    --output code.json

  # Validate generated JSON
  codegov-cli validate --input code.json

  # Apply overrides
  codegov-cli override \
    --original code.json \
    --new code-final.json \
    --overrides overrides.json

Documentation: https://github.com/NSACodeGov/CodeGov`)
}
