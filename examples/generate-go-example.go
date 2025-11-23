package main

import (
	"fmt"
	"log"

	"github.com/NSACodeGov/CodeGov/codegov"
)

func main() {
	// Example 1: Set GitHub OAuth token for higher API rate limits
	// You can also set the OAUTH_TOKEN environment variable instead
	token := "your_40_character_github_token_here"
	if token != "your_40_character_github_token_here" {
		if err := codegov.SetOAuthToken(token); err != nil {
			log.Fatalf("Error setting OAuth token: %v\n", err)
		}
	}

	// Example 2: Verify token is valid
	if !codegov.TestOAuthToken() {
		fmt.Println("Warning: No valid OAuth token found. API rate limits will be reduced.")
	}

	// Example 3: Generate code.gov JSON for multiple organizations
	organizations := []string{"NSACodeGov", "18F"}
	agencyName := "NSA"
	agencyEmail := "opensource@nsa.gov"

	agencyOptions := map[string]string{
		"name":  "NSA Cybersecurity",
		"url":   "https://nsa.gov/contact",
		"phone": "1-800-NSA-CYBER",
	}

	fmt.Println("Generating code.gov JSON inventory...")
	fmt.Printf("Organizations: %v\n", organizations)
	fmt.Printf("Agency: %s\n", agencyName)

	codeGov, err := codegov.NewCodeGovJSON(
		organizations,
		agencyName,
		agencyEmail,
		agencyOptions,
		false, // include private repositories
		false, // include fork repositories
	)
	if err != nil {
		log.Fatalf("Error generating code.gov JSON: %v\n", err)
	}

	fmt.Printf("Generated %d releases\n", len(codeGov.Releases))

	// Example 4: Save to file
	outputPath := "code.json"
	if err := codegov.NewCodeGovJSONFile(
		organizations,
		agencyName,
		agencyEmail,
		agencyOptions,
		false,
		false,
		outputPath,
	); err != nil {
		log.Fatalf("Error saving code.gov JSON: %v\n", err)
	}

	fmt.Printf("Code.gov JSON saved to: %s\n", outputPath)

	// Example 5: Validate the generated JSON
	fmt.Println("\nValidating generated JSON...")
	isValid, errors, err := codegov.TestCodeGovJSONFile(outputPath)
	if err != nil {
		log.Fatalf("Error validating JSON: %v\n", err)
	}

	if isValid {
		fmt.Println("✓ JSON is valid and compliant with code.gov schema v2.0")
	} else {
		fmt.Println("✗ JSON validation errors:")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	// Example 6: Apply overrides from a separate JSON file
	// This allows you to manually update certain fields after generation
	fmt.Println("\nApplying overrides...")
	err = codegov.InvokeCodeGovJsonOverride(
		"code.json",
		"code-final.json",
		"examples/overrides-example.json",
	)
	if err != nil {
		fmt.Printf("Note: Override example not found or error: %v\n", err)
	} else {
		fmt.Println("✓ Overrides applied successfully")

		// Validate the final output
		isValid, errors, _ := codegov.TestCodeGovJSONFile("code-final.json")
		if isValid {
			fmt.Println("✓ Final JSON is valid")
		} else {
			fmt.Println("✗ Final JSON has validation errors:")
			for _, e := range errors {
				fmt.Printf("  - %s\n", e)
			}
		}
	}

	fmt.Println("\nExample complete!")
}
