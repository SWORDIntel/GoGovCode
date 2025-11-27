# GoGovCode – DSMIL Control Plane Framework

GoGovCode is a Go-based control plane framework designed for the **DSMIL (Distributed Secure Multi-level Infrastructure Layer)** ecosystem.

Originally a code.gov inventory generator, GoGovCode has evolved into a **microservice scaffold** that provides:

- **Standardized HTTP server** with health checks, structured logging, and graceful shutdown
- **Hierarchical configuration** supporting dev/test/prod/DSMIL deployment profiles
- **Security-first defaults** with TLS support and audit-ready logging
- **Extensible architecture** ready for PQC (post-quantum cryptography), QUIC protocols, and policy enforcement

GoGovCode serves as the foundation for DSMIL control-plane services, providing consistent patterns for:

- Configuration management across 104+ device types
- Structured event logging with correlation IDs
- Health monitoring and readiness checks
- Authentication and authorization hooks (future phases)
- Audit trail generation and immutable storage (future phases)

## Architecture

GoGovCode follows a clean architecture pattern:

```
cmd/gogovcode/          - Main server entrypoint
api/
  handlers/             - HTTP request handlers
  middleware/           - Request ID, logging, recovery middleware
  routes/               - Route definitions
internal/
  auth/                 - Authentication (future)
  audit/                - Audit logging (future)
  health/               - Health check system
  logging/              - Structured logging with correlation IDs
  policy/               - Policy engine (future)
  server/               - HTTP server with graceful shutdown
  util/                 - Utilities
pkg/
  client/               - Client libraries (future)
  models/               - Data models
  schema/               - JSON schemas
config/                 - Configuration system
```

## Phase 1: Foundation & Baseline Integration

**Status:** ✅ Complete

Phase 1 establishes the foundational microservice scaffold:

- ✅ Project structure aligned with DSMIL expectations
- ✅ Hierarchical config system (file + env + flags)
- ✅ Structured JSON logging with correlation IDs
- ✅ Health endpoints (`/healthz`, `/readyz`)
- ✅ HTTP server with graceful shutdown
- ✅ Base middleware chain (request ID, logging, panic recovery)

## Quick Start

### Running the Server

```bash
# Build
go build -o gogovcode ./cmd/gogovcode

# Run with defaults (dev profile, port 8080)
./gogovcode

# Run with custom profile
./gogovcode -profile prod -port 8443 -tls

# Run with environment variables
export GOGOVCODE_PORT=9000
export GOGOVCODE_LOG_LEVEL=debug
./gogovcode
```

### Health Checks

```bash
# Liveness check (always returns 200)
curl http://localhost:8080/healthz

# Readiness check (checks dependencies)
curl http://localhost:8080/readyz
```

### Configuration

GoGovCode supports hierarchical configuration with priority: **flags > env > file > defaults**

**Configuration file example (`config.json`):**

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "tls": {
    "enabled": true,
    "cert_file": "/path/to/cert.pem",
    "key_file": "/path/to/key.pem"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "service": {
    "name": "gogovcode",
    "version": "1.0.0-phase1"
  },
  "profile": "prod"
}
```

**Environment variables:**

- `GOGOVCODE_HOST` - Server bind host
- `GOGOVCODE_PORT` - Server port
- `GOGOVCODE_LOG_LEVEL` - Log level (debug/info/warn/error)
- `GOGOVCODE_LOG_FORMAT` - Log format (json/text)
- `GOGOVCODE_TLS_ENABLED` - Enable TLS (true/false)
- `GOGOVCODE_TLS_CERT` - TLS certificate path
- `GOGOVCODE_TLS_KEY` - TLS key path

## Legacy CLI Tool

The original code.gov CLI tool is still available at `cmd/codegov-cli/` for generating code inventory JSON files.

```bash
# Build legacy CLI
go build -o codegov-cli ./cmd/codegov-cli

# Generate code.gov JSON
./codegov-cli generate \
  --orgs "NSACodeGov,18F" \
  --agency "NSA" \
  --email "contact@nsa.gov" \
  --output code.json
```

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
