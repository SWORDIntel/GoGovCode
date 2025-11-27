# GoGovCode – Phase 5: SDKs, Device Templates & Identity Integration

**Version:** 1.0  
**Status:** Draft – Ready for LLM refinement  
**Project:** GoGovCode – DSMIL Control Plane Framework  
**Dependencies:** Phase 1–4 complete

---

## 1. Objectives

- Provide standardized **SDKs** in multiple languages.
- Add a **device daemon template generator** for DSMIL devices (0–103).
- Integrate **SPIFFE/SPIRE identity** for strong workload identity and attestation.
- Tighten constant-time guarantees for sensitive crypto/auth code.

---

## 2. Scope

- SDK generation toolchain.  
- Scripted device service template creation.  
- Workload identity integration.  
- Constant-time helper primitives.

---

## 3. Deliverables

1. **Multi-Language SDKs**
   - OpenAPI (or similar) spec derived from GoGovCode APIs.
   - Makefile target:
     - `make generate-sdk`
   - Generated clients:
     - Go
     - Python
     - Rust
     - TypeScript
   - Shared auth & clearance model:
     - helpers for:
       - device tokens
       - clearance fields
       - attaching metadata for audit.

2. **Device Template Generator**
   - Script: `scripts/new-device.sh` (or Go CLI `cmd/gogovcodectl`):
     - Input:
       - device_id
       - layer
       - service name
     - Output:
       - new service folder under `cmd/` or `services/deviceXX/`
       - `main.go` with:
         - server wiring
         - clearance middleware
         - health endpoints
         - bus and/or QUIC hooks
       - default policy file (YAML/JSON)
       - default audit config

3. **Identity Integration (SPIFFE/SPIRE)**
   - `internal/auth/identity.go`:
     - SPIFFE ID support:
       - `spiffe://domain/service/deviceXX`
     - fetches SVID (X.509 or JWT) from SPIRE agent.
     - attaches identity into request context and audit records.
   - Policy tie-in:
     - ability to require certain SPIFFE IDs for:
       - specific routes
       - specific device operations.

4. **Constant-Time Helpers**
   - `internal/util/constant_time.go`:
     - constant-time compare for secrets
     - patterns to avoid secret-dependent branching
   - Documentation on how this aligns with DSLLVM constant-time compilation for adjacent components.

---

## 4. High-Level Tasks

- Define and generate OpenAPI spec from existing Go handlers.
- Integrate an API client generator into the build system.
- Implement `new-device` template generator script/CLI.
- Integrate SPIFFE/SPIRE client libraries.
- Implement constant-time utility helpers and apply them to:
  - auth checks
  - token comparisons
  - crypto-sensitive decision points.

---

## 5. Exit Criteria

- `make generate-sdk` produces SDKs in all targeted languages.
- Running `scripts/new-device.sh 47 7 advanced-aiml` yields:
  - a compilable, runnable device 47 service skeleton
  - route stubs with clearance and audit wired.
- When SPIRE is available:
  - GoGovCode services obtain an SVID and include identity in logs/audit.
- Sensitive comparisons (secrets/tokens) use constant-time utilities, verified by tests or code review.

---

## 6. Overall Outcome

By the end of Phase 5, GoGovCode functions as:

- The **canonical DSMIL control-plane framework**.  
- A generator for 104 device services with consistent security, policy, and audit behavior.  
- A PQC-ready, QUIC-capable, event-bus-aware, identity-enforced substrate that SHRINK, MEMSHADOW, DSAria3, and HURRICANE can all build on.
