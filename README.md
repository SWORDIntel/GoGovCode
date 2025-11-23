# GoGovCode

GoGovCode is a Go-based, cross-platform reimplementation of the NSA Cybersecurity CodeGov module.

Like the original, it creates a [code.gov](https://code.gov/) [code inventory JSON file](https://code.gov/about/compliance/inventory-code) (`code.json`) from GitHub repository information – but without the Windows / PowerShell / .NET prerequisites.

Instead of GAC-installed Newtonsoft dependencies and PowerShell setup, GoGovCode ships as a simple CLI and Go library that you can:

- Run locally on Linux, macOS, or in containers to generate `code.json` on demand  
- Drop into CI/CD pipelines (GitHub Actions, GitLab CI, Jenkins, etc.) to auto-refresh inventories on release or on a schedule  
- Use in platform/compliance workflows to aggregate multiple orgs/repos into a single agency inventory  
- Embed as a Go module in your own internal tools that need to emit or validate code.gov-compliant inventories  

The tool queries GitHub, normalizes repository metadata to the official code.gov schema, and emits a JSON inventory suitable for publication on agency sites, internal catalogs, or further validation with the official schema tools.

## Why code.gov?

For agencies and their contractors, `code.json` isn’t just paperwork – it’s the machine-readable index of what exists, who owns it, and how it can be reused.

Having a clean, code.gov-compliant inventory enables you to:

- Prove policy compliance (e.g., open source release, reuse, licensing) without manual spreadsheets  
- Give program offices, security teams, and auditors a single source of truth for “what code do we actually have?”  
- De-duplicate effort across programs by making existing codebases discoverable instead of re-written  

## How this maps to real-world contractor workflows

GoGovCode is designed to match how modern government-facing teams actually work:

- **Multi-org / multi-tenant setups**  
  Run one pipeline that walks multiple GitHub orgs (agency + integrator + lab) and emits a unified `code.json` per customer, directorate, or classification boundary.

- **Per-engagement inventories**  
  Attach a `code.json` snapshot to each delivery (or release tag) so every engagement has a precise view of what was in-scope at that point in time.

- **Air-gapped and ephemeral runners**  
  Build the binary once, drop it into locked-down CI runners or offline build environments, and regenerate inventories there without needing PowerShell or internet for anything except the Git host you already use.

- **Compliance + security glue**  
  Treat `code.json` as an input to the rest of your stack:  
  - join it with SBOM output  
  - feed it into internal dashboards  
  - use it to drive which repos get extra scanning, telemetry, or hardening.

- **Consulting / oversight roles**  
  When you’re brought in to “make sense of the mess”, pointing GoGovCode at an org and emitting a first-pass `code.json` is a fast way to discover all the moving parts before deeper analysis.

The goal is to make `code.json` generation **cheap enough to do all the time** – locally, in CI, and across multiple organizations – so inventories stop being a painful yearly exercise and become part of the normal delivery pipeline.
