# GoGovCode

GoGovCode is a Go-based, cross-platform reimplementation of the NSA Cybersecurity CodeGov module.

Like the original, it creates a [code.gov](https://code.gov/) [code inventory JSON file](https://code.gov/about/compliance/inventory-code) (`code.json`) from GitHub repository information – but without the Windows / PowerShell / .NET prerequisites.

Instead of GAC-installed Newtonsoft dependencies and PowerShell setup, GoGovCode ships as a simple CLI and Go library that you can:

- Run locally on Linux, macOS, or in containers to generate `code.json` on demand  
- Drop into CI/CD pipelines (GitHub Actions, GitLab CI, Jenkins, etc.) to auto-refresh inventories on release or on a schedule  
- Use in platform/compliance workflows to aggregate multiple orgs/repos into a single agency inventory  
- Embed as a Go module in your own internal tools that need to emit or validate code.gov-compliant inventories  

The tool queries GitHub, normalizes repository metadata to the official code.gov schema, and emits a JSON inventory suitable for publication on agency sites, internal catalogs, or further validation with the [code.gov schema validator](https://code.gov/about/compliance/inventory-code/validate-schema).

## Why code.gov?

The U.S. Federal Source Code Policy (OMB M-16-21) requires agencies to **track and inventory custom-developed software** and to make a portion of that code discoverable and reusable across government – with at least **20% of new custom code released as open source** under the pilot. :contentReference[oaicite:0]{index=0}

To implement this, agencies publish a machine-readable **source code inventory** (typically at `agency.gov/code.json`) following the code.gov metadata schema. :contentReference[oaicite:1]{index=1} This inventory is used to:

- Provide a single, consistent place to discover agency code assets  
- Enable reuse and reduce duplicative software spend across agencies :contentReference[oaicite:2]{index=2}  
- Satisfy internal policy requirements (e.g., DHS, GSA, SEC OSS policies) that reference M-16-21 and code inventories :contentReference[oaicite:3]{index=3}  

GoGovCode exists to make that **inventory-generation step** easy to automate, portable, and CI-friendly, especially for teams that don’t want to maintain Windows/PowerShell environments just to keep `code.json` compliant and up to date.
