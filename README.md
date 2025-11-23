# GoGovCode

GoGovCode is a Go-based, cross-platform reimplementation of the NSA Cybersecurity **CodeGov** module.

Like the original, it creates a [code.gov](https://code.gov/) [code inventory JSON file](https://code.gov/about/compliance/inventory-code) (`code.json`) from GitHub repository information – but without the Windows / PowerShell / .NET prerequisites.

Instead of GAC-installed Newtonsoft dependencies and PowerShell setup, GoGovCode ships as a simple CLI and Go library that you can:

- Run locally on **Linux, macOS, or in containers** to generate `code.json` on demand  
- Drop into **CI/CD pipelines** (GitHub Actions, GitLab CI, Jenkins, etc.) to auto-refresh inventories on release or on a schedule  
- Use in **platform/compliance workflows** to aggregate multiple orgs/repos into a single agency inventory  
- Embed as a **Go module** in your own internal tools that need to emit or validate code.gov-compliant inventories

The tool queries GitHub, normalizes repository metadata to the official code.gov schema, and emits a JSON inventory suitable for publication on agency sites, internal catalogs, or further validation with the [code.gov schema validator](https://code.gov/about/compliance/inventory-code/validate-schema).
