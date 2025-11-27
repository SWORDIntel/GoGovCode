# GoGovCode – Phase 1: Foundation & Baseline Integration

**Version:** 1.0  
**Status:** Draft – Ready for LLM refinement  
**Project:** GoGovCode – DSMIL Control Plane Framework  
**Dependencies:** Existing GoGovCode repo bootstrapped and compiling

---

## 1. Objectives

- Turn GoGovCode into the **standard microservice scaffold** for DSMIL.
- Align project layout, config, logging, and health checks with DSMIL expectations.
- Prepare a clean base for later PQC, clearance, audit, and bus integrations.

---

## 2. Scope

Phase 1 is **non-cryptographic and non-PQC**.  
It focuses on:

- Project structure cleanup  
- Config and environment handling  
- Logging and health endpoints  
- Base HTTP(S) server behavior  
- Basic auth pluggability (no DSMIL rules yet)

---

## 3. Deliverables

1. **Repository Layout & Docs**
   - Updated `README.md` describing DSMIL integration role.
   - High-level architecture diagram for GoGovCode as a DSMIL control-plane framework.
   - Clear folder structure under:
     - `cmd/gogovcode/`
     - `api/handlers`, `api/middleware`, `api/routes`
     - `internal/{auth,audit,health,logging,policy,server,util}`
     - `pkg/{client,models,schema}`

2. **Config System**
   - `config/config.go` with:
     - hierarchical config (file + env + flags)
     - separate “profiles” (dev, test, prod, DSMIL)
   - Strong defaults for:
     - HTTP bind host/port
     - TLS on/off
     - log level
     - Redis/MinIO endpoints (placeholders)

3. **Logging & Health**
   - `internal/logging`:
     - structured logs (JSON)
     - correlation IDs
     - standard fields (service, version, device_id, layer)
   - `internal/health`:
     - `/healthz` (liveness)
     - `/readyz` (readiness)
     - optional status: redis/minio connectivity check stubs

4. **Server Skeleton**
   - `internal/server`:
     - HTTP server with graceful shutdown
     - context-aware request handling
     - base middleware chain:
       - request ID
       - logging
       - panic recovery

---

## 4. High-Level Tasks

- Clean and document repository layout.
- Implement config loader with environment overrides.
- Implement centralized logging package.
- Implement minimal health endpoints.
- Implement standardized server startup/shutdown in `cmd/gogovcode/main.go`.

---

## 5. Exit Criteria

- `go build ./...` passes without errors.
- `go test ./...` passes for foundational packages.
- Running `gogovcode` locally exposes:
  - `/healthz` returning 200
  - `/readyz` returning 200
- Logs are structured JSON and contain:
  - timestamp, level, msg, service, version, request_id
